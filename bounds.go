package main

import (
	"math"
)

// Distance computes the Euclidean distance between two points
func Distance(a, b Vertex) float64 {
	dx, dy, dz := float64(a.X-b.X), float64(a.Y-b.Y), float64(a.Z-b.Z)
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// Find the farthest point from a given reference point
func FarthestPoint(points []Vertex, ref Vertex) Vertex {
	var farthest Vertex
	maxDist := -1.0
	for _, p := range points {
		dist := Distance(ref, p)
		if dist > maxDist {
			maxDist = dist
			farthest = p
		}
	}
	return farthest
}

// Compute bounding sphere using Ritter's Algorithm
func RitterBoundingSphere(points []Vertex) (center Vertex, radius float64) {
	if len(points) == 0 {
		return Vertex{}, 0
	}

	// Step 1: Pick an arbitrary point P0
	p0 := points[0]

	// Step 2: Find P1, the farthest point from P0
	p1 := FarthestPoint(points, p0)

	// Step 3: Find P2, the farthest point from P1
	p2 := FarthestPoint(points, p1)

	// Step 4: Compute initial sphere
	center = Vertex{
		X: (p1.X + p2.X) / 2,
		Y: (p1.Y + p2.Y) / 2,
		Z: (p1.Z + p2.Z) / 2,
		W: 1.0,
		A: 0,
		R: 0,
		G: 0,
		B: 0}
	radius = Distance(p1, p2) / 2

	// Step 5: Expand sphere if needed
	for _, p := range points {
		dist := Distance(center, p)
		if dist > radius {
			// Compute new sphere to include p
			newRadius := (radius + dist) / 2
			ratio := (newRadius - radius) / dist

			center = Vertex{
				X: center.X + float32(float64(p.X-center.X)*ratio),
				Y: center.Y + float32(float64(p.Y-center.Y)*ratio),
				Z: center.Z + float32(float64(p.Z-center.Z)*ratio),
				W: 1.0,
				A: 0,
				R: 0,
				G: 0,
				B: 0}
			radius = newRadius
		}
	}

	return center, radius
}
