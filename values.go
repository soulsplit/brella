package main

// main map to store all the users' holdings
type values map[string]float64

// map currency to a list of values
type valuesMap map[string][]float64

// add data to the map that holds all values from the current session
func (vMap valuesMap) storeValues(currName string, value float64) {
	_, ok := vMap[currName]
	if !ok {
		vMap[currName] = []float64{value}
	} else {
		vMap[currName] = append(vMap[currName], value)
	}
}

// extract only the last price that was recorded already of a given cryptocurrency
func getLastValue(currName string, vMap valuesMap) float64 {
	_, ok := vMap[currName]
	if ok {
		return vMap[currName][len(vMap[currName])-1]
	}
	return 0
}

// extract all the prices that were recorded already of a given cryptocurrency
func getValues(currName string, vMap valuesMap) []float64 {
	_, ok := vMap[currName]
	if ok {
		return vMap[currName]
	}
	return []float64{0.0}
}
