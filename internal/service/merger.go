package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hytale-tools/blockymodel-merger/pkg/blockymodel"
	"github.com/hytale-tools/blockymodel-merger/pkg/character"
	"github.com/hytale-tools/blockymodel-merger/pkg/export"
	"github.com/hytale-tools/blockymodel-merger/pkg/merger"
	"github.com/hytale-tools/blockymodel-merger/pkg/registry"
	"github.com/hytale-tools/blockymodel-merger/pkg/texture"
)

const (
	basePath        = "assets/Characters/Player.blockymodel"
	baseTexturePath = "assets/Characters/Player_Textures/Player_Greyscale.png"
)

// HeadAccessoryEntry extends registry entry with HeadAccessoryType
type HeadAccessoryEntry struct {
	ID                           string `json:"Id"`
	HeadAccessoryType            string `json:"HeadAccessoryType"`
	DisableCharacterPartCategory string `json:"DisableCharacterPartCategory"`
}

// HaircutEntry extends registry entry with HairType
type HaircutEntry struct {
	ID       string `json:"Id"`
	HairType string `json:"HairType"`
}

// MergeService handles character merging operations
type MergeService struct {
	registry         *registry.Registry
	gradientSets     *texture.GradientSets
	baseModel        *blockymodel.BlockyModel
	headAccessories  map[string]HeadAccessoryEntry
	haircuts         map[string]HaircutEntry
	haircutFallbacks map[string]string // HairType -> fallback haircut ID
}

// MergeResult contains the results of a merge operation
type MergeResult struct {
	Model    *blockymodel.BlockyModel
	Atlas    *texture.Atlas
	GLBBytes []byte
}

// NewMergeService creates a new merge service with all required data loaded
func NewMergeService() (*MergeService, error) {
	// Load gradient sets for tinting
	gradientSets, err := texture.LoadGradientSets()
	if err != nil {
		return nil, fmt.Errorf("loading gradient sets: %w", err)
	}

	// Load accessory registry
	reg, err := registry.Load()
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}

	// Load base player model
	baseModel, err := blockymodel.Load(basePath)
	if err != nil {
		return nil, fmt.Errorf("loading base model: %w", err)
	}

	// Load head accessories for HeadAccessoryType
	headAccessories, err := loadHeadAccessories("data/HeadAccessory.json")
	if err != nil {
		return nil, fmt.Errorf("loading head accessories: %w", err)
	}

	// Load haircuts for HairType
	haircuts, err := loadHaircuts("data/Haircuts.json")
	if err != nil {
		return nil, fmt.Errorf("loading haircuts: %w", err)
	}

	// Load haircut fallbacks
	haircutFallbacks, err := loadHaircutFallbacks("data/HaircutFallbacks.json")
	if err != nil {
		return nil, fmt.Errorf("loading haircut fallbacks: %w", err)
	}

	return &MergeService{
		registry:         reg,
		gradientSets:     gradientSets,
		baseModel:        baseModel,
		headAccessories:  headAccessories,
		haircuts:         haircuts,
		haircutFallbacks: haircutFallbacks,
	}, nil
}

// MergeFromJSON merges a character from JSON data and returns the result
func (s *MergeService) MergeFromJSON(charJSON []byte) (*MergeResult, error) {
	// Parse character data
	var charData character.CharacterData
	if err := json.Unmarshal(charJSON, &charData); err != nil {
		return nil, fmt.Errorf("parsing character JSON: %w", err)
	}

	// Apply haircut fallback if headAccessory requires it
	s.applyHaircutFallback(&charData)

	// Resolve accessories
	result, err := charData.ResolveAccessories(s.registry)
	if err != nil {
		return nil, fmt.Errorf("resolving accessories: %w", err)
	}

	// Create merger from base model
	m, err := merger.New(s.baseModel)
	if err != nil {
		return nil, fmt.Errorf("creating merger: %w", err)
	}

	// Merge each accessory
	for _, acc := range result.Accessories {
		accessory, err := blockymodel.Load(acc.Path)
		if err != nil {
			return nil, fmt.Errorf("loading accessory %s: %w", acc.Path, err)
		}

		if err := m.Merge(accessory, acc.Spec.ID); err != nil {
			return nil, fmt.Errorf("merging accessory %s: %w", acc.Path, err)
		}
	}

	// Get merged model
	mergedModel := m.Result()

	// Process textures
	var tintedTextures []*texture.TintedTexture
	skinTone := charData.GetSkinTone()

	// Load and tint base player texture
	if skinTone != "" {
		baseTinted, err := texture.ProcessAccessoryTexture(
			"_base",
			baseTexturePath,
			"Skin",
			skinTone,
			s.gradientSets,
		)
		if err == nil {
			tintedTextures = append(tintedTextures, baseTinted)
		}
	} else {
		baseImg, err := texture.LoadImage(baseTexturePath)
		if err == nil {
			baseTex := &texture.TintedTexture{
				Name:         "_base",
				Image:        baseImg,
				OriginalPath: baseTexturePath,
			}
			tintedTextures = append(tintedTextures, baseTex)
		}
	}

	// Process accessory textures
	for _, acc := range result.Accessories {
		if acc.ResolvedTexture == nil {
			continue
		}

		var tinted *texture.TintedTexture

		if acc.ResolvedTexture.DirectTexture != "" {
			img, err := texture.LoadImage(acc.ResolvedTexture.DirectTexture)
			if err != nil {
				continue
			}
			tinted = &texture.TintedTexture{
				Name:         acc.Spec.ID,
				Image:        img,
				OriginalPath: acc.ResolvedTexture.DirectTexture,
			}
		} else if acc.ResolvedTexture.GreyscaleTexture != "" {
			var err error
			tinted, err = texture.ProcessAccessoryTexture(
				acc.Spec.ID,
				acc.ResolvedTexture.GreyscaleTexture,
				acc.ResolvedTexture.GradientSet,
				acc.Spec.Color,
				s.gradientSets,
			)
			if err != nil {
				continue
			}
		} else {
			continue
		}

		tintedTextures = append(tintedTextures, tinted)
	}

	// Pack textures into atlas
	var atlas *texture.Atlas
	if len(tintedTextures) > 0 {
		var err error
		atlas, err = texture.PackAtlasSimple(tintedTextures, 1)
		if err != nil {
			return nil, fmt.Errorf("packing atlas: %w", err)
		}

		// Update texture offsets in the merged model
		for _, tex := range tintedTextures {
			if tex.Name == "_base" {
				continue
			}

			x, y, _, _, ok := atlas.GetPixelCoords(tex.Name)
			if !ok {
				continue
			}

			// Find all node IDs that came from this accessory
			nodeIDs := make(map[string]bool)
			for nodeID, accessoryID := range m.NodeSources {
				if accessoryID == tex.Name {
					nodeIDs[nodeID] = true
				}
			}

			if len(nodeIDs) > 0 {
				offset := blockymodel.AtlasOffset{X: float64(x), Y: float64(y)}
				blockymodel.UpdateTextureOffsets(mergedModel.Nodes, nodeIDs, offset)
			}
		}
	}

	// Export to GLB
	exporter := export.NewGLBExporter()

	var materialIdx uint32
	if atlas != nil {
		w, h := atlas.Image.Bounds().Dx(), atlas.Image.Bounds().Dy()
		exporter.SetAtlasSize(float64(w), float64(h))

		atlasBytes, err := texture.EncodePNG(atlas.Image)
		if err != nil {
			return nil, fmt.Errorf("encoding atlas: %w", err)
		}

		texIdx := exporter.AddTexture(atlasBytes)
		materialIdx = exporter.AddMaterial("textured", texIdx)
	} else {
		exporter.SetAtlasSize(64, 64)
		materialIdx = 0
	}

	if err := exporter.ExportModel(mergedModel, materialIdx); err != nil {
		return nil, fmt.Errorf("exporting model: %w", err)
	}

	glbBytes, err := exporter.Bytes()
	if err != nil {
		return nil, fmt.Errorf("getting GLB bytes: %w", err)
	}

	return &MergeResult{
		Model:    mergedModel,
		Atlas:    atlas,
		GLBBytes: glbBytes,
	}, nil
}

// applyHaircutFallback modifies haircut based on headAccessory type
func (s *MergeService) applyHaircutFallback(charData *character.CharacterData) {
	if charData.HeadAccessory == nil || *charData.HeadAccessory == "" {
		return
	}

	// Parse head accessory ID
	headAccID := strings.Split(*charData.HeadAccessory, ".")[0]
	headAcc, ok := s.headAccessories[headAccID]
	if !ok {
		return
	}

	// Check if headAccessory disables haircut entirely
	if headAcc.DisableCharacterPartCategory == "Haircut" {
		charData.Haircut = nil
		return
	}

	// Check headAccessory type
	switch headAcc.HeadAccessoryType {
	case "FullyCovering":
		// No hair visible
		charData.Haircut = nil
	case "HalfCovering":
		// Use fallback hairstyle
		if charData.Haircut != nil && *charData.Haircut != "" {
			s.setFallbackHaircut(charData)
		}
	}
	// "Simple" or empty: keep original haircut
}

// setFallbackHaircut replaces haircut with appropriate fallback based on HairType
func (s *MergeService) setFallbackHaircut(charData *character.CharacterData) {
	if charData.Haircut == nil || *charData.Haircut == "" {
		return
	}

	// Parse haircut spec (ID.Color.Variant)
	parts := strings.Split(*charData.Haircut, ".")
	haircutID := parts[0]
	color := ""
	if len(parts) > 1 {
		color = parts[1]
	}

	// Get haircut entry to find HairType
	haircut, ok := s.haircuts[haircutID]
	if !ok {
		return
	}

	// Get fallback haircut ID for this HairType
	fallbackID, ok := s.haircutFallbacks[haircut.HairType]
	if !ok {
		return
	}

	// Build new haircut string with fallback ID but same color
	newHaircut := fallbackID
	if color != "" {
		newHaircut = fallbackID + "." + color
	}
	charData.Haircut = &newHaircut
}

// loadHeadAccessories loads head accessory data from JSON file
func loadHeadAccessories(path string) (map[string]HeadAccessoryEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []HeadAccessoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	result := make(map[string]HeadAccessoryEntry)
	for _, e := range entries {
		if e.ID != "" {
			result[e.ID] = e
		}
	}
	return result, nil
}

// loadHaircuts loads haircut data from JSON file
func loadHaircuts(path string) (map[string]HaircutEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []HaircutEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	result := make(map[string]HaircutEntry)
	for _, e := range entries {
		if e.ID != "" {
			result[e.ID] = e
		}
	}
	return result, nil
}

// loadHaircutFallbacks loads haircut fallback mappings from JSON file
func loadHaircutFallbacks(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
