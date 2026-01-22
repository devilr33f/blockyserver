package api

// OpenAPISpec is the OpenAPI 3.0 specification for the BlockyServer API
const OpenAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "BlockyModel Merger API",
    "description": "HTTP API for rendering Hytale character models as GLB, PNG, or animated GIF.\n\n**BlockyServer** by [devilreef](https://github.com/devilr33f)\n\n**BlockyModel Merger** by [JackGamesFTW](https://github.com/JackGamesFTW)",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "/",
      "description": "Current server"
    }
  ],
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check",
        "operationId": "getHealth",
        "tags": ["System"],
        "responses": {
          "200": {
            "description": "Server is healthy",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": {
                      "type": "string",
                      "example": "ok"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/render/glb": {
      "post": {
        "summary": "Render character as GLB",
        "description": "Renders a character and returns GLB binary file.",
        "operationId": "renderGLB",
        "tags": ["Render"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/CharacterConfig"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "GLB binary file",
            "content": {
              "model/gltf-binary": {
                "schema": {
                  "type": "string",
                  "format": "binary"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "500": {
            "$ref": "#/components/responses/InternalError"
          }
        }
      }
    },
    "/render/png": {
      "post": {
        "summary": "Render character as PNG",
        "description": "Renders a character as a PNG image with configurable rotation and background.",
        "operationId": "renderPNG",
        "tags": ["Render"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/PNGRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "PNG image",
            "content": {
              "image/png": {
                "schema": {
                  "type": "string",
                  "format": "binary"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "500": {
            "$ref": "#/components/responses/InternalError"
          }
        }
      }
    },
    "/render/gif": {
      "post": {
        "summary": "Render character as animated GIF",
        "description": "Renders a character as an animated rotating GIF.",
        "operationId": "renderGIF",
        "tags": ["Render"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/GIFRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Animated GIF",
            "content": {
              "image/gif": {
                "schema": {
                  "type": "string",
                  "format": "binary"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "500": {
            "$ref": "#/components/responses/InternalError"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "CharacterConfig": {
        "type": "object",
        "description": "Character appearance configuration. All fields are optional. Format: \"AccessoryId.Color.Variant\"",
        "properties": {
          "bodyCharacteristic": {"type": "string", "example": "Default.02", "description": "Body type and skin tone"},
          "underwear": {"type": "string", "example": "Underwear_Male", "description": "Base underwear"},
          "face": {"type": "string", "example": "Face_A", "description": "Face shape"},
          "ears": {"type": "string", "example": "Ears_A", "description": "Ear type"},
          "mouth": {"type": "string", "example": "Mouth_A", "description": "Mouth shape"},
          "haircut": {"type": "string", "example": "Scavenger_Hair.PitchBlack", "description": "Hair style and color"},
          "facialHair": {"type": "string", "example": "Beard_A.Brown", "description": "Beard/mustache"},
          "eyebrows": {"type": "string", "example": "Eyebrows_A.Black", "description": "Eyebrow style and color"},
          "eyes": {"type": "string", "example": "Large_Eyes.Pink", "description": "Eye style and color"},
          "pants": {"type": "string", "example": "Pants_A.Blue", "description": "Lower body clothing"},
          "overpants": {"type": "string", "description": "Pants overlay (belt, etc.)"},
          "undertop": {"type": "string", "example": "Shirt_A.White", "description": "Shirt/undershirt"},
          "overtop": {"type": "string", "example": "Jacket_A.Red", "description": "Jacket/coat"},
          "shoes": {"type": "string", "example": "Boots_A.Brown", "description": "Footwear"},
          "headAccessory": {"type": "string", "description": "Hat/helmet"},
          "faceAccessory": {"type": "string", "description": "Glasses/mask"},
          "earAccessory": {"type": "string", "description": "Earrings"},
          "skinFeature": {"type": "string", "description": "Tattoos/markings"},
          "gloves": {"type": "string", "description": "Hand accessories"},
          "cape": {"type": "string", "description": "Back cape"}
        }
      },
      "PNGRequest": {
        "type": "object",
        "required": ["character"],
        "properties": {
          "character": {"$ref": "#/components/schemas/CharacterConfig"},
          "rotation": {"type": "number", "default": 0, "description": "Rotation in degrees"},
          "background": {"type": "string", "default": "transparent", "description": "\"transparent\" or hex color \"#RRGGBB\""},
          "width": {"type": "integer", "default": 512, "description": "Image width in pixels"},
          "height": {"type": "integer", "default": 512, "description": "Image height in pixels"}
        }
      },
      "GIFRequest": {
        "type": "object",
        "required": ["character"],
        "properties": {
          "character": {"$ref": "#/components/schemas/CharacterConfig"},
          "background": {"type": "string", "default": "#FFFFFF", "description": "Hex color (no transparency for GIF)"},
          "frames": {"type": "integer", "default": 36, "description": "Number of frames (36 = 10° per frame)"},
          "width": {"type": "integer", "default": 512, "description": "Image width in pixels"},
          "height": {"type": "integer", "default": 512, "description": "Image height in pixels"},
          "delay": {"type": "integer", "default": 5, "description": "Centiseconds between frames"},
          "dithering": {"type": "boolean", "default": true, "description": "Enable Floyd-Steinberg dithering (disable for faster rendering)"}
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "error": {"type": "string", "description": "Error message"}
        }
      }
    },
    "responses": {
      "BadRequest": {
        "description": "Bad request",
        "content": {
          "application/json": {
            "schema": {"$ref": "#/components/schemas/ErrorResponse"}
          }
        }
      },
      "InternalError": {
        "description": "Internal server error",
        "content": {
          "application/json": {
            "schema": {"$ref": "#/components/schemas/ErrorResponse"}
          }
        }
      }
    }
  }
}`

// SwaggerUIHTML returns the Swagger UI HTML page
const SwaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>BlockyServer API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`
