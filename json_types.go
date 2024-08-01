package http_client

import (
	"encoding/json"
	"strconv"
	"time"
)

type Date struct {
	time.Time
}

func (d Date) MarshalSchema() string {
	return d.Time.Format("2006-01-02")
}

type DateTime struct {
	time.Time
}

func (d DateTime) MarshalSchema() string {
	return d.Time.Format(time.RFC3339)
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return json.Marshal(nil)
	}

	return json.Marshal(d.Time.Format("2006-01-02 15:04:05"))
}

func (d *DateTime) UnmarshalJSON(text []byte) (err error) {
	var value string
	err = json.Unmarshal(text, &value)
	if err != nil {
		return err
	}

	if value == "" {
		return nil
	}

	d.Time, err = time.Parse("2006-01-02 15:04:05", value)
	if err == nil {
		return nil
	}

	// lastly try standard date
	d.Time, err = time.Parse(time.RFC3339, value)
	return err
}

type StringFloat float64

func (f *StringFloat) UnmarshalJSON(text []byte) (err error) {
	var flt float64
	err = json.Unmarshal(text, &flt)
	if err == nil {
		*f = StringFloat(flt)
		return err
	}

	// error, so try string
	var s string
	err = json.Unmarshal(text, &s)
	if err != nil {
		return err
	}

	flt, err = strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	*f = StringFloat(flt)
	return nil
}

type IntString string

func (f *IntString) UnmarshalJSON(text []byte) (err error) {
	var str string
	err = json.Unmarshal(text, &str)
	if err == nil {
		*f = IntString(str)
		return err
	}

	// error, so try int
	var integer int
	err = json.Unmarshal(text, &integer)
	if err != nil {
		return err
	}

	str = strconv.Itoa(integer)
	*f = IntString(str)
	return nil
}
