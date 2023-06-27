package lib

import (
	"fmt"
	"math"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/umahmood/haversine"
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

func CalcHaversine(lat1 float64, long1 float64, lat2 float64, long2 float64) float64 {
	location1 := haversine.Coord{Lat: lat2, Lon: long2}
	location2 := haversine.Coord{Lat: lat1, Lon: long1}

	fmt.Println("location1: ", location1)
	fmt.Println("location2: ", location2)

	mi, km := haversine.Distance(location1, location2)

	fmt.Printf("Miles: %f\n", mi)
	fmt.Printf("Kilometers: %f\n", km)

	return mi
}

func CalcTextSimilarity(text1 string, text2 string) float64 {
	fmt.Printf("text 1 %s\n", text1)
	fmt.Printf("text 2 %s\n", text2)
	return strutil.Similarity(text1, text2, metrics.NewJaccard())
}
