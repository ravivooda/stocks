package models

import "strconv"

type TwoRoundedFloat float64

func (t TwoRoundedFloat) MarshalJSON() ([]byte, error) {
	if float64(t) == float64(int(t)) {
		return []byte(strconv.FormatFloat(float64(t), 'f', 1, 32)), nil
	}
	return []byte(strconv.FormatFloat(float64(t), 'f', -1, 32)), nil
}
