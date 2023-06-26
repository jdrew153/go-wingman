package lib

import (
	"math"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

func GetDistanceFromCoords(lat1 float64, long1 float64, lat2 float64, long2 float64) float64 {
	lat1 = math.Pi * lat1 / 180
	lat2 = math.Pi * lat2 / 180
	long1 = math.Pi * long1 / 180
	long2 = math.Pi * long2 / 180

	dlon := long2 - long1
	dlat := lat2 - lat1

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	earthRadius := 6371.0

	distance := earthRadius * c

	return distance

}

func CalcTextSimilarity(text1 string, text2 string) float64 {
	return strutil.Similarity(text1, text2, metrics.NewHamming())
}
