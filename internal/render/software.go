package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/fogleman/fauxgl"
	"github.com/qmuntal/gltf"
)

// GLBToMesh converts GLB bytes to a fauxgl mesh with texture
func GLBToMesh(glbBytes []byte, atlasImage image.Image) (*fauxgl.Mesh, error) {
	doc := new(gltf.Document)
	if err := gltf.NewDecoder(bytes.NewReader(glbBytes)).Decode(doc); err != nil {
		return nil, fmt.Errorf("parsing GLB: %w", err)
	}

	mesh := fauxgl.NewEmptyMesh()

	// Process all nodes in the scene
	if len(doc.Scenes) == 0 || len(doc.Scenes[0].Nodes) == 0 {
		return nil, fmt.Errorf("GLB has no scene nodes")
	}

	// Build node transforms
	for _, nodeIdx := range doc.Scenes[0].Nodes {
		if err := processNode(doc, int(nodeIdx), fauxgl.Identity(), mesh, atlasImage); err != nil {
			return nil, err
		}
	}

	return mesh, nil
}

func processNode(doc *gltf.Document, nodeIdx int, parentTransform fauxgl.Matrix, mesh *fauxgl.Mesh, atlasImage image.Image) error {
	node := doc.Nodes[nodeIdx]

	// Build local transform in TRS order: Translation * Rotation * Scale
	localTransform := fauxgl.Identity()

	// Apply scale first (rightmost in matrix multiplication)
	if node.Scale != [3]float64{1, 1, 1} && node.Scale != [3]float64{0, 0, 0} {
		localTransform = fauxgl.Scale(fauxgl.V(node.Scale[0], node.Scale[1], node.Scale[2])).Mul(localTransform)
	}

	// Apply rotation
	if node.Rotation != [4]float64{0, 0, 0, 1} {
		qx, qy, qz, qw := node.Rotation[0], node.Rotation[1], node.Rotation[2], node.Rotation[3]
		R := quaternionToMatrix(qx, qy, qz, qw)
		localTransform = R.Mul(localTransform)
	}

	// Apply translation last (leftmost)
	if node.Translation != [3]float64{0, 0, 0} {
		localTransform = fauxgl.Translate(fauxgl.V(node.Translation[0], node.Translation[1], node.Translation[2])).Mul(localTransform)
	}

	// World transform = parent * local
	worldTransform := parentTransform.Mul(localTransform)

	// Process mesh if present
	if node.Mesh != nil {
		gltfMesh := doc.Meshes[*node.Mesh]
		for _, prim := range gltfMesh.Primitives {
			if err := processPrimitive(doc, prim, worldTransform, mesh, atlasImage); err != nil {
				return err
			}
		}
	}

	// Process children
	for _, childIdx := range node.Children {
		if err := processNode(doc, int(childIdx), worldTransform, mesh, atlasImage); err != nil {
			return err
		}
	}

	return nil
}

func processPrimitive(doc *gltf.Document, prim *gltf.Primitive, transform fauxgl.Matrix, mesh *fauxgl.Mesh, atlasImage image.Image) error {
	// Get position accessor
	posAccessorIdx, ok := prim.Attributes[gltf.POSITION]
	if !ok {
		return nil
	}
	posAccessor := doc.Accessors[posAccessorIdx]
	positions, err := readVec3Accessor(doc, posAccessor)
	if err != nil {
		return fmt.Errorf("reading positions: %w", err)
	}

	// Get UV accessor (optional)
	var uvs [][2]float32
	if uvAccessorIdx, ok := prim.Attributes[gltf.TEXCOORD_0]; ok {
		uvAccessor := doc.Accessors[uvAccessorIdx]
		uvs, err = readVec2Accessor(doc, uvAccessor)
		if err != nil {
			return fmt.Errorf("reading UVs: %w", err)
		}
	}

	// Get indices
	var indices []uint32
	if prim.Indices != nil {
		indicesAccessor := doc.Accessors[*prim.Indices]
		indices, err = readIndicesAccessor(doc, indicesAccessor)
		if err != nil {
			return fmt.Errorf("reading indices: %w", err)
		}
	} else {
		for i := 0; i < len(positions); i++ {
			indices = append(indices, uint32(i))
		}
	}

	// Build triangles
	for i := 0; i < len(indices); i += 3 {
		i0, i1, i2 := indices[i], indices[i+1], indices[i+2]

		p0 := positions[i0]
		p1 := positions[i1]
		p2 := positions[i2]

		v0 := fauxgl.V(float64(p0[0]), float64(p0[1]), float64(p0[2]))
		v1 := fauxgl.V(float64(p1[0]), float64(p1[1]), float64(p1[2]))
		v2 := fauxgl.V(float64(p2[0]), float64(p2[1]), float64(p2[2]))

		// Transform positions
		v0 = transform.MulPosition(v0)
		v1 = transform.MulPosition(v1)
		v2 = transform.MulPosition(v2)

		tri := fauxgl.Triangle{
			V1: fauxgl.Vertex{Position: v0, Color: fauxgl.White},
			V2: fauxgl.Vertex{Position: v1, Color: fauxgl.White},
			V3: fauxgl.Vertex{Position: v2, Color: fauxgl.White},
		}

		// Add UVs if available
		if len(uvs) > 0 {
			uv0 := uvs[i0]
			uv1 := uvs[i1]
			uv2 := uvs[i2]

			tri.V1.Texture = fauxgl.V(float64(uv0[0]), 1.0-float64(uv0[1]), 0)
			tri.V2.Texture = fauxgl.V(float64(uv1[0]), 1.0-float64(uv1[1]), 0)
			tri.V3.Texture = fauxgl.V(float64(uv2[0]), 1.0-float64(uv2[1]), 0)
		}

		// Compute normal from transformed vertices
		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)
		normal := edge1.Cross(edge2).Normalize()
		tri.V1.Normal = normal
		tri.V2.Normal = normal
		tri.V3.Normal = normal

		mesh.Triangles = append(mesh.Triangles, &tri)
	}

	return nil
}

func readVec3Accessor(doc *gltf.Document, accessor *gltf.Accessor) ([][3]float32, error) {
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor has no buffer view")
	}

	bv := doc.BufferViews[*accessor.BufferView]
	buf := doc.Buffers[bv.Buffer]

	data := buf.Data[bv.ByteOffset+accessor.ByteOffset:]
	stride := bv.ByteStride
	if stride == 0 {
		stride = 12
	}

	count := int(accessor.Count)
	result := make([][3]float32, count)
	for i := 0; i < count; i++ {
		offset := i * int(stride)
		result[i][0] = readFloat32LE(data[offset:])
		result[i][1] = readFloat32LE(data[offset+4:])
		result[i][2] = readFloat32LE(data[offset+8:])
	}

	return result, nil
}

func readVec2Accessor(doc *gltf.Document, accessor *gltf.Accessor) ([][2]float32, error) {
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor has no buffer view")
	}

	bv := doc.BufferViews[*accessor.BufferView]
	buf := doc.Buffers[bv.Buffer]

	data := buf.Data[bv.ByteOffset+accessor.ByteOffset:]
	stride := bv.ByteStride
	if stride == 0 {
		stride = 8
	}

	count := int(accessor.Count)
	result := make([][2]float32, count)
	for i := 0; i < count; i++ {
		offset := i * int(stride)
		result[i][0] = readFloat32LE(data[offset:])
		result[i][1] = readFloat32LE(data[offset+4:])
	}

	return result, nil
}

func readIndicesAccessor(doc *gltf.Document, accessor *gltf.Accessor) ([]uint32, error) {
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor has no buffer view")
	}

	bv := doc.BufferViews[*accessor.BufferView]
	buf := doc.Buffers[bv.Buffer]

	data := buf.Data[bv.ByteOffset+accessor.ByteOffset:]
	count := int(accessor.Count)
	result := make([]uint32, count)

	switch accessor.ComponentType {
	case gltf.ComponentUbyte:
		for i := 0; i < count; i++ {
			result[i] = uint32(data[i])
		}
	case gltf.ComponentUshort:
		for i := 0; i < count; i++ {
			result[i] = uint32(readUint16LE(data[i*2:]))
		}
	case gltf.ComponentUint:
		for i := 0; i < count; i++ {
			result[i] = readUint32LE(data[i*4:])
		}
	default:
		return nil, fmt.Errorf("unsupported index component type: %v", accessor.ComponentType)
	}

	return result, nil
}

func readFloat32LE(data []byte) float32 {
	bits := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	return math.Float32frombits(bits)
}

func readUint16LE(data []byte) uint16 {
	return uint16(data[0]) | uint16(data[1])<<8
}

func readUint32LE(data []byte) uint32 {
	return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
}

// AlphaTestShader is a texture shader with alpha cutoff
type AlphaTestShader struct {
	Matrix      fauxgl.Matrix
	Texture     fauxgl.Texture
	AlphaCutoff float64
}

func NewAlphaTestShader(matrix fauxgl.Matrix, texture fauxgl.Texture, alphaCutoff float64) *AlphaTestShader {
	return &AlphaTestShader{matrix, texture, alphaCutoff}
}

func (s *AlphaTestShader) Vertex(v fauxgl.Vertex) fauxgl.Vertex {
	v.Output = s.Matrix.MulPositionW(v.Position)
	return v
}

func (s *AlphaTestShader) Fragment(v fauxgl.Vertex) fauxgl.Color {
	color := s.Texture.Sample(v.Texture.X, v.Texture.Y)
	if color.A < s.AlphaCutoff {
		return fauxgl.Discard
	}
	return color
}

func quaternionToMatrix(x, y, z, w float64) fauxgl.Matrix {
	n := math.Sqrt(x*x + y*y + z*z + w*w)
	if n > 0 {
		x /= n
		y /= n
		z /= n
		w /= n
	}

	xx := x * x
	yy := y * y
	zz := z * z
	xy := x * y
	xz := x * z
	yz := y * z
	wx := w * x
	wy := w * y
	wz := w * z

	return fauxgl.Matrix{
		1 - 2*(yy+zz), 2 * (xy - wz), 2 * (xz + wy), 0,
		2 * (xy + wz), 1 - 2*(xx+zz), 2 * (yz - wx), 0,
		2 * (xz - wy), 2 * (yz + wx), 1 - 2*(xx+yy), 0,
		0, 0, 0, 1,
	}
}

// RenderScene renders a mesh with the given parameters
func RenderScene(mesh *fauxgl.Mesh, atlasImage image.Image, rotationY float64, width, height int, bgColor color.Color) image.Image {
	context := fauxgl.NewContext(width, height)
	context.Cull = fauxgl.CullNone
	context.AlphaBlend = false

	r, g, b, a := bgColor.RGBA()
	if a == 0 {
		context.ClearColor = fauxgl.Transparent
	} else {
		context.ClearColor = fauxgl.Color{
			R: float64(r) / 65535.0,
			G: float64(g) / 65535.0,
			B: float64(b) / 65535.0,
			A: float64(a) / 65535.0,
		}
	}
	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	box := mesh.BoundingBox()
	modelCenter := box.Center()
	modelSize := box.Size()

	aspect := float64(width) / float64(height)
	fovy := 30.0
	near := 0.1
	far := 100.0

	maxDim := math.Max(modelSize.X, math.Max(modelSize.Y, modelSize.Z))
	cameraDistance := maxDim / (2 * math.Tan(fauxgl.Radians(fovy/2))) * 1.5

	eye := fauxgl.V(modelCenter.X, modelCenter.Y, modelCenter.Z+cameraDistance)
	center := modelCenter
	up := fauxgl.V(0, 1, 0)

	modelMatrix := fauxgl.Rotate(fauxgl.V(0, 1, 0), fauxgl.Radians(rotationY))
	viewMatrix := fauxgl.LookAt(eye, center, up)
	projMatrix := fauxgl.Perspective(fovy, aspect, near, far)
	matrix := projMatrix.Mul(viewMatrix).Mul(modelMatrix)

	var shader fauxgl.Shader
	if atlasImage != nil {
		shader = NewAlphaTestShader(matrix, fauxgl.NewImageTexture(atlasImage), 0.05)
	} else {
		shader = fauxgl.NewSolidColorShader(matrix, fauxgl.HexColor("#CCCCCC"))
	}

	context.Shader = shader
	context.DrawMesh(mesh)

	return context.Image()
}

// ParseHexColor parses a hex color string like "#RRGGBB" or "transparent"
func ParseHexColor(hex string) (color.Color, error) {
	if hex == "transparent" || hex == "" {
		return color.RGBA{0, 0, 0, 0}, nil
	}

	if len(hex) != 7 || hex[0] != '#' {
		return nil, fmt.Errorf("invalid hex color: %s (expected #RRGGBB)", hex)
	}

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return nil, fmt.Errorf("parsing hex color %s: %w", hex, err)
	}

	return color.RGBA{r, g, b, 255}, nil
}
