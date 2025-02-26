package main

type Vertex struct {
	X, Y, Z, W float32
	A, R, G, B float32
	flushed    bool
}

type Normal struct {
	X, Y, Z, W float32
	flushed    bool
}

type TextureCoord struct {
	U, V, W float32
	flushed bool
}

type Tangent struct {
	tan, bitan Normal
	flushed    bool
}

type BoundSphere struct {
	center Vertex
	radius float32
}

const FindVertexScore_CacheDecayPower float32 = 1.5
const FindVertexScore_LastTriScore float32 = 0.75
const FindVertexScore_ValenceBoostScale float32 = 2.0
const FindVertexScore_ValenceBoostPower float32 = 0.5

type Face struct {
	edges        uint8
	v            []uint32
	n            []uint32
	t            []uint32
	uv           []uint32
	materialID   uint32
	materialName string
	mortonCode   uint32
	complete     bool
}

const ILLUM0 uint32 = 0   // Color on and Ambient off
const ILLUM1 uint32 = 1   // Color on and Ambient on
const ILLUM2 uint32 = 2   // Highlight on
const ILLUM3 uint32 = 3   // Reflection on and Ray trace on
const ILLUM4 uint32 = 4   // Transparency: Glass on, Reflection: Ray trace on
const ILLUM5 uint32 = 5   // Reflection: Fresnel on and Ray trace on
const ILLUM6 uint32 = 6   // Transparency: Refraction on, Reflection: Fresnel off and Ray trace on
const ILLUM7 uint32 = 7   // Transparency: Refraction on, Reflection: Fresnel on and Ray trace on
const ILLUM8 uint32 = 8   // Reflection on and Ray trace off
const ILLUM9 uint32 = 9   // Transparency: Glass on, Reflection: Ray trace off
const ILLUM10 uint32 = 10 // Casts shadows onto invisible surfaces

type Material struct {
	name                string
	diffuse             [3]float32
	specular            [3]float32
	ambient             [3]float32
	transmissive        [3]float32
	emissive            [3]float32
	power               float32
	transparency        float32
	refractivity        float32
	illum               uint32
	roughness           float32
	metallic            float32
	sheen               float32
	clearcoat_thickness float32
	clearcoat_roughness float32
	aniso               float32
	aniso_rotation      float32
	texture             string
}
