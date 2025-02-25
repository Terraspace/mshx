**WaveFront OBJ to MSHX Converter Utility**

**MSHX Format**

    magic: char[4] ; 'MSHX'
    version:       uint32       ; version of the MSHX file (currently 1)
    vertexCount:   uint32
    normalCount:   uint32
    tangentCount:  uint32
    uvCount:       uint32
    faceCount:     uint32
    materialCount: uint32
    vertexType:    uint32       ; 0=xyz, 1=xyzargb
    
    boundingSphere: x,y,z,radius (float)
    ; A close-to-optimal bounding sphere is generated for the mesh
    
    vertices[vertexCount]:
    x,y,z,<a,r,g,b> (float,<float>) ; w assumed = 1.0, if no color data in OBJ file, <ARGB> is omitted
    
    normals[normalCount]:
    nx,ny,nz (float) ; w assumed = 0.0
    ; *** Normals from an OBJ file are re-normalized on conversion
    
    tangents[tangentCount]:
    tux,tuy,tuz (float) ; w assumed = 0.0 (tangent)
    tvx,tvy,tvz (float) ; w assumed = 0.0 (bitangent)
    
    uvs[uvCount]:
    u,v (float)
    
    faces[faceCount]:
    edge-count (uint8)            ; [3=tri, 4=quad...]
    v1,v2,v3,[v4]... (uint32)     ; index into above vertex buffer [mandatory]
    n1,n2,n3,[n4]... (uint32)     ; index into above normal buffer [skipped if no normals above]
    t1,t2,t3,[t4]... (uint32)     ; index into above tangent buffer [skipped if no tangets above]
    uv1,uv2,uv3,[uv4]... (uint32) ; index into above uv buffer [skipped if no uvs above]
    materialID (uint32)           ; mandatory = 0 if no materials
    ; for a quad, the 4 vertices are tested to ensure they are coplanar and convex
    ; face winding order is assumed to be correct in the source OBJ file
    ; vertex/face reordering pre-pass in converter to optimize for vertex cache
    ; *** source OBJ files must use absolute and not relative indices
    ; *** OBJ files are assumed to only support triangle and quad, not higher order polyongs
    
    materials[materialCount]:
    ambientColor (argb[] float32)      ; ambient colour
    diffuseColor (argb[] float32)      ; diffuse colour
    specularColor (argb[] float32)     ; specular colour
    emissiveColor (argb[] float32)     ; emissive colour
    transmissiveColor (argb[] float32) ; transmissive colour
    specularPower (float32)            ; specular power
    emissive (float32)
    roughness (float32)                ; mix between diffuse and specular (fixed for an entire material) [0.0 - 1.0] [diffuse - specular] - used when no roughness map provided below.
    metal (float32)                    ; [0.0 - 1.0] how reflective a material is / 0=nonmetal, 1=metal
    mode (uint32)
    sheen (float32)
    clearcoat thickness (float32)
    clearcore roughness (float32)
    anisotropy (float32)
    aniostropy rotation (float32)
    texture map string length (uint32)
    texture map name (byte[])


