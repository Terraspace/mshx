package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

var curMaterialName string
var curMaterialIdx uint32 = 0

var vertices []Vertex
var normals []Normal
var textureCoords []TextureCoord
var faces []Face
var materials []Material
var materialMap map[string]uint32 = make(map[string]uint32)
var boundSphere BoundSphere

var vertexType uint32 = 0

var dPtr *bool
var moPtr *bool
var qPtr *int
var lePtr *bool
var bePtr *bool
var silentPtr *bool
var inputFileName string
var outputFileName string

func ParseCommandLine() bool {
	fmt.Println("-- OBJ file converter v0.1 --")

	// Get Command Line flags.
	lePtr = flag.Bool("le", false, "Output data as little endian")
	bePtr = flag.Bool("be", false, "Output data as big endian")
	silentPtr = flag.Bool("silent", false, "Do not output any messages")
	moPtr = flag.Bool("mo", false, "Optimise mesh data")
	dPtr = flag.Bool("d", false, "Remove duplicate vertices/normals/uvs")
	qPtr = flag.Int("q", 0, "0=No quad validation, 1=Validate quad faces and fail on error, 2=Validate quad faces and convert degenrate quads to triangles, 3=Convert all quad faces to triangles")
	flag.Parse()

	// Handle endianness flags.
	if *lePtr && *bePtr {
		fmt.Println("Error: Cannot specify both little and big endian.")
		return false
	} else if !*lePtr && !*bePtr {
		fmt.Println("Warning: No endianness specified. Defaulting to little endian.")
		*lePtr = true
	}

	// Get command line arguments for input and output file.
	var args []string = flag.Args() //os.Args[1:]
	var argCount int = len(args)

	if argCount < 2 {
		fmt.Println("Usage: objconv [flags] <input file> <output file>")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		return false
	}

	inputFileName = args[0]
	outputFileName = args[1]
	return true
}

func GenerateBoundingSphere() {
	center, radius := RitterBoundingSphere(vertices)
	boundSphere.center = center
	boundSphere.radius = float32(radius)
	fmt.Printf("Generated Bounding Sphere: %v\n", boundSphere)
}

func ProcessMaterialFile(materialFileName string) error {

	// Open material file.
	materialFile, err := os.Open(materialFileName)
	if err != nil {
		fmt.Printf("Error opening material file %s: %v\n", materialFileName, err)
		return err
	}
	defer materialFile.Close()

	var inMaterial bool = false
	var materialName string
	var material Material

	var scanner *bufio.Scanner = bufio.NewScanner(materialFile)
	for scanner.Scan() {
		var line string = strings.Trim(scanner.Text(), " \t")
		// If the line begins with # or is empty, skip it.
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		lineParts := strings.Split(line, " ")
		switch lineParts[0] {
		case "newmtl":
			inMaterial = true
			materialName = lineParts[1]
			material = *new(Material)
			material.name = materialName
			materials = append(materials, material)
			materialMap[materialName] = uint32(len(materials) - 1)
			if !*silentPtr {
				fmt.Printf("Defining Material %s\n", materialName)
			}
		case "Kd":
			if inMaterial {
				var r, g, b float32
				fmt.Sscanf(line, "Kd %f %f %f", &r, &g, &b)
				fmt.Printf("Diffuse: %f %f %f\n", r, g, b)
				materials[len(materials)-1].diffuse = [3]float32{r, g, b}
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Ke":
			if inMaterial {
				var r, g, b float32
				fmt.Sscanf(line, "Ke %f %f %f", &r, &g, &b)
				fmt.Printf("Emissive: %f %f %f\n", r, g, b)
				materials[len(materials)-1].emissive = [3]float32{r, g, b}
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Ka":
			if inMaterial {
				var r, g, b float32
				fmt.Sscanf(line, "Ka %f %f %f", &r, &g, &b)
				fmt.Printf("Ambient: %f %f %f\n", r, g, b)
				materials[len(materials)-1].ambient = [3]float32{r, g, b}
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Ks":
			if inMaterial {
				var r, g, b float32
				fmt.Sscanf(line, "Ks %f %f %f", &r, &g, &b)
				fmt.Printf("Specular: %f %f %f\n", r, g, b)
				materials[len(materials)-1].specular = [3]float32{r, g, b}
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Tf":
			if inMaterial {
				if lineParts[1] == "xyz" || lineParts[1] == "spectral" {
					fmt.Printf("Error: Only RGB transmissive colours supported and no spectral transmission data supported.\n")
					return errors.New("spectral transmission data not supported")
				}
				var r, g, b float32
				fmt.Sscanf(line, "Tf %f %f %f", &r, &g, &b)
				fmt.Printf("Transmissive: %f %f %f\n", r, g, b)
				materials[len(materials)-1].transmissive = [3]float32{r, g, b}
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Ns":
			if inMaterial {
				var power float32
				fmt.Sscanf(line, "Ns %f", &power)
				fmt.Printf("Specular Power: %f\n", power)
				materials[len(materials)-1].power = power
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "d":
			if inMaterial {
				var t float32
				fmt.Sscanf(line, "d %f", &t)
				fmt.Printf("Dissolve: %f\n", t)
				materials[len(materials)-1].transparency = 1.0 - t
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Tr":
			if inMaterial {
				var t float32
				fmt.Sscanf(line, "Tr %f", &t)
				fmt.Printf("Transparency: %f\n", t)
				materials[len(materials)-1].transparency = t
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Ni":
			if inMaterial {
				var t float32
				fmt.Sscanf(line, "Ni %f", &t)
				fmt.Printf("Refractivity: %f\n", t)
				materials[len(materials)-1].refractivity = t
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "illum":
			if inMaterial {
				var i uint32
				fmt.Sscanf(line, "illum %d", &i)
				fmt.Printf("Illumination Mode: %d\n", i)
				materials[len(materials)-1].illum = i
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Pr":
			if inMaterial {
				var r float32
				fmt.Sscanf(line, "Pr %f", &r)
				fmt.Printf("Roughness: %f\n", r)
				materials[len(materials)-1].roughness = r
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Pm":
			if inMaterial {
				var m float32
				fmt.Sscanf(line, "Pm %f", &m)
				fmt.Printf("Metallic: %f\n", m)
				materials[len(materials)-1].metallic = m
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Ps":
			if inMaterial {
				var m float32
				fmt.Sscanf(line, "Ps %f", &m)
				fmt.Printf("Sheen: %f\n", m)
				materials[len(materials)-1].sheen = m
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Pc":
			if inMaterial {
				var m float32
				fmt.Sscanf(line, "Pc %f", &m)
				fmt.Printf("Clearcoat Thickness: %f\n", m)
				materials[len(materials)-1].clearcoat_thickness = m
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "Pcr":
			if inMaterial {
				var m float32
				fmt.Sscanf(line, "Pcr %f", &m)
				fmt.Printf("Metallic: %f\n", m)
				materials[len(materials)-1].clearcoat_roughness = m
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "aniso":
			if inMaterial {
				var m float32
				fmt.Sscanf(line, "aniso %f", &m)
				fmt.Printf("Anisotropy: %f\n", m)
				materials[len(materials)-1].aniso = m
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "anisor":
			if inMaterial {
				var m float32
				fmt.Sscanf(line, "anisor %f", &m)
				fmt.Printf("Anisotropy: %f\n", m)
				materials[len(materials)-1].aniso_rotation = m
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		case "map_Kd":
			if inMaterial {
				var txt string
				fmt.Sscanf(line, "map_Kd %s", &txt)
				fmt.Printf("Texture Map: %s\n", txt)
				materials[len(materials)-1].texture = txt
			} else {
				fmt.Printf("Error: Material properties defined outside of material block.\n")
				return errors.New("material properties defined outside of material block")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", materialFileName, err)
		return err
	}

	return nil
}

func ProcessOBJFile(inputFile *os.File) error {
	// Read input file line by line.
	var scanner *bufio.Scanner = bufio.NewScanner(inputFile)
	for scanner.Scan() {
		var line string = strings.Trim(scanner.Text(), " \t")

		// If the line begins with # or is empty, skip it.
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Split the line into tokens, and decide how to handle each line
		// based on the first token which identifies the type of data on that line.
		lineParts := strings.Split(line, " ")
		switch lineParts[0] {
		case "v":
			var vertex Vertex = Vertex{0.0, 0.0, 0.0, 1.0, 1.0, 1.0, 1.0, 1.0, false}
			if len(lineParts) == 4 {
				fmt.Sscanf(line, "v %f %f %f", &vertex.X, &vertex.Y, &vertex.Z)
			} else if len(lineParts) == 5 {
				fmt.Sscanf(line, "v %f %f %f %f", &vertex.X, &vertex.Y, &vertex.Z, &vertex.W)
			} else if len(lineParts) == 7 {
				vertexType = 1
				fmt.Sscanf(line, "v %f %f %f %f %f %f %f", &vertex.X, &vertex.Y, &vertex.Z, &vertex.R, &vertex.G, &vertex.B)
			}
			vertices = append(vertices, vertex)
			if !*silentPtr {
				fmt.Printf("Vertex %v\n", vertex)
			}
		case "vt":
			var textureCoord TextureCoord
			textureCoord.flushed = false
			if len(lineParts) == 2 {
				fmt.Sscanf(line, "vt %f", &textureCoord.U)
				textureCoord.V = 0.0
			} else if len(lineParts) == 3 {
				fmt.Sscanf(line, "vt %f %f", &textureCoord.U, &textureCoord.V)
			} else if len(lineParts) == 4 {
				fmt.Sscanf(line, "vt %f %f %f", &textureCoord.U, &textureCoord.V)
			}
			textureCoords = append(textureCoords, textureCoord)
			if !*silentPtr {
				fmt.Printf("TextureCoord %v\n", textureCoord)
			}
		case "vn":
			var normal Normal
			normal.flushed = false
			fmt.Sscanf(line, "vn %f %f %f", &normal.X, &normal.Y, &normal.Z)
			normal.W = 0.0
			normal.normalize()
			normals = append(normals, normal)
			if !*silentPtr {
				fmt.Printf("Normal %v\n", normal)
			}
		case "usemtl":
			curMaterialName = lineParts[1]
			if !*silentPtr {
				fmt.Printf("Using Material %s\n", curMaterialName)
			}
		case "mtllib":
			err := ProcessMaterialFile(lineParts[1])
			if err != nil {
				fmt.Printf("Error processing material file: %v\n", err)
				return err
			}
		case "f":
			var face Face
			face.complete = false
			if len(lineParts) == 4 {
				face.edges = 3
			} else if len(lineParts) == 5 {
				face.edges = 4
			} else {
				fmt.Println("Error: Only triangles and quads are supported.")
				return errors.New("invalid face type")
			}
			for i := 1; i < len(lineParts); i++ {
				vertParts := strings.Split(lineParts[i], "/")
				if len(vertParts) >= 1 {
					idx, err := strconv.Atoi(vertParts[0])
					if err != nil {
						return fmt.Errorf("invalid vertex index: %v", err)
					}
					face.v = append(face.v, uint32(idx)-1)
				}
				if len(vertParts) >= 2 {
					idx, err := strconv.Atoi(vertParts[1])
					if err != nil {
						return fmt.Errorf("invalid texture index: %v", err)
					}
					face.uv = append(face.uv, uint32(idx)-1)
				}
				if len(vertParts) == 3 {
					idx, err := strconv.Atoi(vertParts[2])
					if err != nil {
						return fmt.Errorf("invalid normal index: %v", err)
					}
					face.n = append(face.n, uint32(idx)-1)
				}
				if len(vertParts) > 3 {
					return errors.New("invalid vertex index format on face")
				}
			}
			face.materialName = curMaterialName
			faces = append(faces, face)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", inputFileName, err)
		return err
	}

	return nil
}

// Cross product of two 3D vectors
func crossProduct(ax, ay, az, bx, by, bz float64) (float64, float64, float64) {
	return ay*bz - az*by, az*bx - ax*bz, ax*by - ay*bx
}

// Dot product of two 3D vectors
func dotProduct(ax, ay, az, bx, by, bz float64) float64 {
	return ax*bx + ay*by + az*bz
}

func (n *Normal) normalize() {
	var length float32 = float32(math.Sqrt(float64(n.X*n.X + n.Y*n.Y + n.Z*n.Z)))
	n.X /= length
	n.Y /= length
	n.Z /= length
}

// Cross product of two 2D vectors (returns only the z-component)
func crossProductZ(ax, ay, bx, by float64) float64 {
	return ax*by - ay*bx
}

// Check if 4 points form a convex quadrilateral
func isConvex(ax, ay, bx, by, cx, cy, dx, dy float64) bool {
	// Compute edge vectors
	v1x, v1y := bx-ax, by-ay
	v2x, v2y := cx-bx, cy-by
	v3x, v3y := dx-cx, dy-cy
	v4x, v4y := ax-dx, ay-dy

	// Compute cross products (only the z-component)
	c1 := crossProductZ(v1x, v1y, v2x, v2y)
	c2 := crossProductZ(v2x, v2y, v3x, v3y)
	c3 := crossProductZ(v3x, v3y, v4x, v4y)
	c4 := crossProductZ(v4x, v4y, v1x, v1y)

	// All cross products must have the same sign
	return (c1 > 0 && c2 > 0 && c3 > 0 && c4 > 0) || (c1 < 0 && c2 < 0 && c3 < 0 && c4 < 0)
}

func (f *Face) ValidateQuad() error {
	var abx float64 = float64(vertices[f.v[1]].X - vertices[f.v[0]].X)
	var aby float64 = float64(vertices[f.v[1]].Y - vertices[f.v[0]].Y)
	var abz float64 = float64(vertices[f.v[1]].Z - vertices[f.v[0]].Z)
	var acx float64 = float64(vertices[f.v[2]].X - vertices[f.v[0]].X)
	var acy float64 = float64(vertices[f.v[2]].Y - vertices[f.v[0]].Y)
	var acz float64 = float64(vertices[f.v[2]].Z - vertices[f.v[0]].Z)
	var adx float64 = float64(vertices[f.v[3]].X - vertices[f.v[0]].X)
	var ady float64 = float64(vertices[f.v[3]].Y - vertices[f.v[0]].Y)
	var adz float64 = float64(vertices[f.v[3]].Z - vertices[f.v[0]].Z)

	// Compute cross product AB × AC and AC × AD
	nx1, ny1, nz1 := crossProduct(abx, aby, abz, acx, acy, acz)
	nx2, ny2, nz2 := crossProduct(acx, acy, acz, adx, ady, adz)
	len1 := math.Sqrt(nx1*nx1 + ny1*ny1 + nz1*nz1)
	len2 := math.Sqrt(nx2*nx2 + ny2*ny2 + nz2*nz2)
	nx1 /= len1
	ny1 /= len1
	nz1 /= len1
	nx2 /= len2
	ny2 /= len2
	nz2 /= len2

	// Compute dot product (AB × AC) • (AC x AD)
	dot := dotProduct(nx1, ny1, nz1, nx2, ny2, nz2)

	if math.Abs(dot) < 0.999 {
		fmt.Printf("Quad face is not planar: %v\n", dot)
		fmt.Printf("%f %f %f %f %f %f %f %f %f %f %f %f\n", vertices[f.v[0]].X, vertices[f.v[0]].Y, vertices[f.v[0]].Z,
			vertices[f.v[1]].X, vertices[f.v[1]].Y, vertices[f.v[1]].Z,
			vertices[f.v[2]].X, vertices[f.v[2]].Y, vertices[f.v[2]].Z,
			vertices[f.v[3]].X, vertices[f.v[3]].Y, vertices[f.v[3]].Z)
		return errors.New("quad face is not planar")
	}

	if !isConvex(float64(vertices[f.v[0]].X), float64(vertices[f.v[0]].Y),
		float64(vertices[f.v[1]].X), float64(vertices[f.v[1]].Y),
		float64(vertices[f.v[2]].X), float64(vertices[f.v[2]].Y),
		float64(vertices[f.v[3]].X), float64(vertices[f.v[3]].Y)) {
		fmt.Printf("Quad face is not convex: %v\n", f)
		return errors.New("quad face is not convex")
	}

	return nil
}

func RemoveAtIndex[T any](s []T, index int) []T {
	return slices.Delete(s, index, index+1) // Remove element at index
}

func ConvertQuadToTriangles(f *Face) {
	f.edges = 3
	//0,1,2 - 0,2,3
	var newFace Face
	newFace.edges = 3
	newFace.materialID = f.materialID
	newFace.materialName = f.materialName
	newFace.n = make([]uint32, 3)
	newFace.uv = make([]uint32, 3)
	newFace.v = make([]uint32, 3)
	newFace.t = make([]uint32, 0)
	newFace.v[0] = f.v[0]
	newFace.v[1] = f.v[2]
	newFace.v[2] = f.v[3]
	newFace.n[0] = f.n[0]
	newFace.n[1] = f.n[2]
	newFace.n[2] = f.n[3]
	newFace.uv[0] = f.uv[0]
	newFace.uv[1] = f.uv[2]
	newFace.uv[2] = f.uv[3]
	if len(f.t) == 4 {
		newFace.t[0] = f.t[0]
		newFace.t[1] = f.t[2]
		newFace.t[2] = f.t[3]
	}
	faces = append(faces, newFace)
	f.v = RemoveAtIndex(f.v, 3)
	f.n = RemoveAtIndex(f.n, 3)
	f.uv = RemoveAtIndex(f.uv, 3)
	if len(f.t) == 4 {
		f.t = RemoveAtIndex(f.t, 3)
	}
}

func FakeQuadCheck(f *Face) {
	// Check for a quad face that is actually a triangle.
	var vertexUse map[uint32]int = make(map[uint32]int)
	vertexUse[f.v[0]]++
	vertexUse[f.v[1]]++
	vertexUse[f.v[2]]++
	vertexUse[f.v[3]]++
	if len(vertexUse) == 3 {
		fmt.Printf("Fake Quad - Triangle face found: %v\n", f)
	}
}

func Morton3D(x, y, z uint32) uint32 {
	return interleaveBits(x) | (interleaveBits(y) << 1) | (interleaveBits(z) << 2)
}

func interleaveBits(x uint32) uint32 {
	x = (x | (x << 16)) & 0x030000FF
	x = (x | (x << 8)) & 0x0300F00F
	x = (x | (x << 4)) & 0x030C30C3
	x = (x | (x << 2)) & 0x09249249
	return x
}

func DeDupe(vT, nT, uvT float64) {

	var dupeV int = 0
	var dupeN int = 0
	var dupeU int = 0

	// Vertices
	for i := 0; i < len(vertices); i++ {
		if !vertices[i].flushed {
			continue
		}
		for j := i + 1; j < len(vertices); j++ {
			dx := vertices[i].X - vertices[j].X
			dy := vertices[i].Y - vertices[j].Y
			dz := vertices[i].Z - vertices[j].Z
			d := math.Sqrt(float64(dx*dx + dy*dy + dz*dz))
			if d < vT {
				for k := 0; k < len(faces); k++ {
					for l := 0; l < int(faces[k].edges); l++ {
						if faces[k].v[l] == uint32(j) {
							faces[k].v[l] = uint32(i)
						}
					}
				}
				vertices[j].flushed = true
				dupeV++
			}
		}
	}
	for i := 0; i < len(vertices); i++ {
		if vertices[i].flushed {
			vertices = RemoveAtIndex(vertices, i)
			for j := 0; j < len(faces); j++ {
				for l := 0; l < int(faces[j].edges); l++ {
					if faces[j].v[l] > uint32(i) {
						faces[j].v[l]--
					}
				}
			}
			i--
		}
	}

	// Normals
	for i := 0; i < len(normals); i++ {
		if normals[i].flushed {
			continue
		}
		for j := i + 1; j < len(normals); j++ {
			dx := math.Abs(float64(normals[i].X - normals[j].X))
			dy := math.Abs(float64(normals[i].Y - normals[j].Y))
			dz := math.Abs(float64(normals[i].Z - normals[j].Z))
			if dx < nT && dy < nT && dz < nT {
				for k := 0; k < len(faces); k++ {
					for l := 0; l < int(faces[k].edges); l++ {
						if faces[k].n[l] == uint32(j) {
							faces[k].n[l] = uint32(i)
						}
					}
				}
				normals[j].flushed = true
				dupeN++
			}
		}
	}
	for i := 0; i < len(normals); i++ {
		if normals[i].flushed {
			normals = RemoveAtIndex(normals, i)
			for j := 0; j < len(faces); j++ {
				for l := 0; l < int(faces[j].edges); l++ {
					if faces[j].n[l] > uint32(i) {
						faces[j].n[l]--
					}
				}
			}
			i--
		}
	}

	// UVS
	for i := 0; i < len(textureCoords); i++ {
		if textureCoords[i].flushed {
			continue
		}
		for j := i + 1; j < len(textureCoords); j++ {
			du := math.Abs(float64(textureCoords[i].U - textureCoords[j].U))
			dv := math.Abs(float64(textureCoords[i].V - textureCoords[j].V))
			if du < uvT && dv < uvT {
				for k := 0; k < len(faces); k++ {
					for l := 0; l < int(faces[k].edges); l++ {
						if faces[k].uv[l] == uint32(j) {
							faces[k].uv[l] = uint32(i)
						}
					}
				}
				textureCoords[j].flushed = true
				dupeU++
			}
		}
	}
	for i := 0; i < len(textureCoords); i++ {
		if textureCoords[i].flushed {
			textureCoords = RemoveAtIndex(textureCoords, i)
			for j := 0; j < len(faces); j++ {
				for l := 0; l < int(faces[j].edges); l++ {
					if faces[j].uv[l] > uint32(i) {
						faces[j].uv[l]--
					}
				}
			}
			i--
		}
	}

	fmt.Printf("Removed %d duplicate vertices.\n", dupeV)
	fmt.Printf("Removed %d duplicate normals.\n", dupeN)
	fmt.Printf("Removed %d duplicate texture coords.\n", dupeU)

}

func OptimiseMesh() {
	// Use the bounding sphere to define a conservative spatial extent for the mesh
	var extents = [6]float32{boundSphere.center.X - boundSphere.radius, boundSphere.center.Y - boundSphere.radius, boundSphere.center.Z - boundSphere.radius,
		boundSphere.center.X + boundSphere.radius, boundSphere.center.Y + boundSphere.radius, boundSphere.center.Z + boundSphere.radius}

	for i := 0; i < len(faces); i++ {
		// Find the centroid of the face
		var cx float32 = 0.0
		var cy float32 = 0.0
		var cz float32 = 0.0
		for j := 0; j < int(faces[i].edges); j++ {
			cx += vertices[faces[i].v[j]].X
			cy += vertices[faces[i].v[j]].Y
			cz += vertices[faces[i].v[j]].Z
		}
		cx /= float32(faces[i].edges)
		cy /= float32(faces[i].edges)
		cz /= float32(faces[i].edges)

		// Normalize the centroid to [0.0 - 1.0] within the bounding sphere range
		cx = (cx - extents[0]) / (extents[3] - extents[0])
		cy = (cy - extents[1]) / (extents[4] - extents[1])
		cz = (cz - extents[2]) / (extents[5] - extents[2])

		// Quantize the normalized value in the range [0 - 1024]
		var icx uint32 = uint32(cx * 1024.0)
		var icy uint32 = uint32(cy * 1024.0)
		var icz uint32 = uint32(cz * 1024.0)

		faces[i].mortonCode = Morton3D(icx, icy, icz)
	}

	// Sort the faces based on their Morton Code
	slices.SortFunc(faces, func(a, b Face) int {
		if a.mortonCode < b.mortonCode {
			return -1
		} else if a.mortonCode > b.mortonCode {
			return 1
		}
		return 0
	})

	// Setup which faces have been processed
	for i := 0; i < len(faces); i++ {
		faces[i].complete = false
	}
	// Create a lookup of all the faces that use a particular vertex for faster
	// updating when we need to remap the vertex indices.
	vertexFaceUse := make([][]uint32, len(vertices))
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			vidx := faces[i].v[j]
			vertexFaceUse[vidx] = append(vertexFaceUse[vidx], uint32(i))
		}
	}
	// Remap face->vertex references
	var newVertices = []Vertex{}
	var curIndex = 0
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			vidx := faces[i].v[j]
			if !vertices[vidx].flushed {
				for k := 0; k < len(vertexFaceUse[vidx]); k++ {
					face := faces[vertexFaceUse[vidx][k]]
					if !face.complete {
						for l := 0; l < int(face.edges); l++ {
							if face.v[l] == vidx {
								face.v[l] = uint32(curIndex)
							}
						}
					}
				}
				vertexFaceUse[vidx] = []uint32{}
				newVertices = append(newVertices, vertices[vidx])
				vertices[vidx].flushed = true
				curIndex++
			}
		}
		faces[i].complete = true
	}
	vertices = newVertices

	// Setup which faces have been processed
	for i := 0; i < len(faces); i++ {
		faces[i].complete = false
	}
	// Create a lookup of all the faces that use a particular vertex for faster
	// updating when we need to remap the vertex indices.
	normalFaceUse := make([][]uint32, len(normals))
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			nidx := faces[i].n[j]
			normalFaceUse[nidx] = append(normalFaceUse[nidx], uint32(i))
		}
	}
	// Remap face->normal references
	var newNormals = []Normal{}
	curIndex = 0
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			nidx := faces[i].n[j]
			if !normals[nidx].flushed {
				for k := 0; k < len(normalFaceUse[nidx]); k++ {
					face := faces[normalFaceUse[nidx][k]]
					if !face.complete {
						for l := 0; l < int(face.edges); l++ {
							if face.n[l] == nidx {
								face.n[l] = uint32(curIndex)
							}
						}
					}
				}
				normalFaceUse[nidx] = []uint32{}
				newNormals = append(newNormals, normals[nidx])
				normals[nidx].flushed = true
				curIndex++
			}
		}
		faces[i].complete = true
	}
	normals = newNormals

	// Setup which faces have been processed
	for i := 0; i < len(faces); i++ {
		faces[i].complete = false
	}
	// Create a lookup of all the faces that use a particular vertex for faster
	// updating when we need to remap the vertex indices.
	uvFaceUse := make([][]uint32, len(textureCoords))
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			tidx := faces[i].uv[j]
			uvFaceUse[tidx] = append(uvFaceUse[tidx], uint32(i))
		}
	}
	// Remap face->uv references
	var newTextureCoords = []TextureCoord{}
	curIndex = 0
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			tidx := faces[i].uv[j]
			if !textureCoords[tidx].flushed {
				for k := 0; k < len(uvFaceUse[tidx]); k++ {
					face := faces[uvFaceUse[tidx][k]]
					if !face.complete {
						for l := 0; l < int(face.edges); l++ {
							if face.uv[l] == tidx {
								face.uv[l] = uint32(curIndex)
							}
						}
					}
				}
				uvFaceUse[tidx] = []uint32{}
				newTextureCoords = append(newTextureCoords, textureCoords[tidx])
				textureCoords[tidx].flushed = true
				curIndex++
			}
		}
		faces[i].complete = true
	}
	textureCoords = newTextureCoords
}

func main() {
	var err error
	var inputFile *os.File
	var outputFile *os.File

	cmdResult := ParseCommandLine()
	if !cmdResult {
		return
	}

	// Open input and output files.
	inputFile, err = os.Open(inputFileName)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", inputFileName, err)
		return
	}
	defer inputFile.Close()

	outputFile, err = os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outputFileName, err)
		return
	}
	defer outputFile.Close()

	// Parse in the OBJ file.
	err = ProcessOBJFile(inputFile)
	if err != nil {
		return
	}

	// Validate Quad Face Structure.
	var i int = 0
	for i < len(faces) {

		if !*silentPtr {
			fmt.Printf("Processing Face %d\n", i)
		}

		if faces[i].edges == 4 && *qPtr != 3 {
			FakeQuadCheck(&faces[i])
		}

		// Cmd line option, force all quads to triangle conversion
		if faces[i].edges == 4 && *qPtr == 3 {
			ConvertQuadToTriangles(&faces[i])
		} else if faces[i].edges == 4 && *qPtr > 0 {
			err = faces[i].ValidateQuad()
			if err != nil {
				if *qPtr == 1 {
					fmt.Printf("Error validating quad face: %v\n", err)
					return
				} else if *qPtr == 2 {
					// Convert quad face to triangles.
					fmt.Printf("Invalid quad found - converting to triangles..")
					ConvertQuadToTriangles(&faces[i])
					fmt.Printf("[ok]\n")
				}
			} else {
				if !*silentPtr {
					fmt.Printf("[ok]\n")
				}
			}
		}

		i++
	}

	// Process material names to index values.
	if !*silentPtr {
		fmt.Println("Faces before mesh optimsation:")
	}
	for i := range faces {
		faces[i].materialID = materialMap[faces[i].materialName]
		if !*silentPtr {
			fmt.Println(faces[i])
		}
	}

	// Generate the bounding sphere.
	GenerateBoundingSphere()

	// If required, de-dupe vertices, uvs and normals
	if *dPtr {
		DeDupe(0.0001, 0.00001, 0.00001)
	}

	var totalErr int = 0
	var curIdx int = int(faces[0].v[0])
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			totalErr += int(math.Abs(float64(int(faces[i].v[j]) - curIdx)))
			curIdx = int(faces[i].v[j])
		}
	}
	fmt.Println("Total vertex stride distance: ", totalErr)

	// Optimize the mesh data.
	if *moPtr {
		if !*silentPtr {
			fmt.Println("Faces after mesh optimsation:")
		}
		OptimiseMesh()
		if !*silentPtr {
			for i := range faces {
				fmt.Println(faces[i])
			}
		}
	}

	totalErr = 0
	curIdx = int(faces[0].v[0])
	for i := 0; i < len(faces); i++ {
		for j := 0; j < int(faces[i].edges); j++ {
			totalErr += int(math.Abs(float64(int(faces[i].v[j]) - curIdx)))
			curIdx = int(faces[i].v[j])
		}
	}
	fmt.Println("Total vertex stride distance: ", totalErr)

	// Write the output file.
	fmt.Println("Writing output file...")
	WriteOutput(outputFile)
	fmt.Println("Done.")
}

func WriteOutput(outputFile *os.File) {
	writer := bufio.NewWriter(outputFile)

	// Choose the byte order based on the flags
	var byteOrder binary.ByteOrder
	if *lePtr {
		byteOrder = binary.LittleEndian
	} else {
		byteOrder = binary.BigEndian
	}

	binary.Write(writer, byteOrder, []byte("MSHX"))             // Magic header
	binary.Write(writer, byteOrder, uint32(1))                  // Version number
	binary.Write(writer, byteOrder, uint32(len(vertices)))      // Number of vertices
	binary.Write(writer, byteOrder, uint32(len(normals)))       // Number of normals
	binary.Write(writer, byteOrder, uint32(0))                  // Number of tangent vectors
	binary.Write(writer, byteOrder, uint32(len(textureCoords))) // Number of texture coordinates
	binary.Write(writer, byteOrder, uint32(len(faces)))         // Number of faces
	binary.Write(writer, byteOrder, uint32(len(materials)))     // Number of materials

	binary.Write(writer, byteOrder, vertexType)

	binary.Write(writer, byteOrder, boundSphere.center.X)
	binary.Write(writer, byteOrder, boundSphere.center.Y)
	binary.Write(writer, byteOrder, boundSphere.center.Z)
	binary.Write(writer, byteOrder, boundSphere.radius)

	for i := 0; i < len(vertices); i++ {
		binary.Write(writer, byteOrder, vertices[i].X)
		binary.Write(writer, byteOrder, vertices[i].Y)
		binary.Write(writer, byteOrder, vertices[i].Z)
		if vertexType == 1 {
			binary.Write(writer, byteOrder, vertices[i].A)
			binary.Write(writer, byteOrder, vertices[i].R)
			binary.Write(writer, byteOrder, vertices[i].G)
			binary.Write(writer, byteOrder, vertices[i].B)
		}
	}

	for i := 0; i < len(normals); i++ {
		binary.Write(writer, byteOrder, normals[i].X)
		binary.Write(writer, byteOrder, normals[i].Y)
		binary.Write(writer, byteOrder, normals[i].Z)
	}

	for i := 0; i < len(textureCoords); i++ {
		binary.Write(writer, byteOrder, textureCoords[i].U)
		binary.Write(writer, byteOrder, textureCoords[i].V)
	}

	for i := 0; i < len(faces); i++ {
		binary.Write(writer, byteOrder, faces[i].edges)
		for j := 0; j < int(faces[i].edges); j++ {
			binary.Write(writer, byteOrder, faces[i].v[j])
		}
		for j := 0; j < int(faces[i].edges); j++ {
			binary.Write(writer, byteOrder, faces[i].n[j])
		}
		for j := 0; j < int(faces[i].edges); j++ {
			binary.Write(writer, byteOrder, faces[i].uv[j])
		}
		binary.Write(writer, byteOrder, faces[i].materialID)
	}

	for i := 0; i < len(materials); i++ {
		binary.Write(writer, byteOrder, materials[i].diffuse)
		binary.Write(writer, byteOrder, materials[i].specular)
		binary.Write(writer, byteOrder, materials[i].ambient)
		binary.Write(writer, byteOrder, materials[i].transmissive)
		binary.Write(writer, byteOrder, materials[i].emissive)
		binary.Write(writer, byteOrder, materials[i].power)
		binary.Write(writer, byteOrder, materials[i].transparency)
		binary.Write(writer, byteOrder, materials[i].refractivity)
		binary.Write(writer, byteOrder, materials[i].illum)
		binary.Write(writer, byteOrder, materials[i].roughness)
		binary.Write(writer, byteOrder, materials[i].metallic)
		binary.Write(writer, byteOrder, materials[i].sheen)
		binary.Write(writer, byteOrder, materials[i].clearcoat_thickness)
		binary.Write(writer, byteOrder, materials[i].clearcoat_roughness)
		binary.Write(writer, byteOrder, materials[i].aniso)
		binary.Write(writer, byteOrder, materials[i].aniso_rotation)
		binary.Write(writer, byteOrder, uint32(len(materials[i].texture)))
		writer.WriteString(materials[i].texture)
	}

	// Flush the writer to ensure all data is written to the file
	if err := writer.Flush(); err != nil {
		fmt.Printf("Error flushing writer: %v\n", err)
	}
}
