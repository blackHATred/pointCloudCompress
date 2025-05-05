package compress

import "math"

type Point struct {
	X, Y, Z, Intensity float32
}

// VoxelGridFilter выполняет downsampling с шагом `leafSize`.
func VoxelGridFilter(points []Point, leafSize float32) []Point {
	voxelMap := make(map[[3]int]Point)

	for _, p := range points {
		ix := int(math.Floor(float64(p.X / leafSize)))
		iy := int(math.Floor(float64(p.Y / leafSize)))
		iz := int(math.Floor(float64(p.Z / leafSize)))
		key := [3]int{ix, iy, iz}
		// Просто оставляем первую точку в вокселе
		if _, ok := voxelMap[key]; !ok {
			voxelMap[key] = p
		}
	}

	filtered := make([]Point, 0, len(voxelMap))
	for _, p := range voxelMap {
		filtered = append(filtered, p)
	}
	return filtered
}
