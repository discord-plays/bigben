package utils

import "github.com/disgoorg/snowflake/v2"

type XormSnowflake snowflake.ID

func (x *XormSnowflake) FromDB(data []byte) error {
	a, err := snowflake.Parse(string(data))
	if err != nil {
		return err
	}
	*x = XormSnowflake(a)
	return nil
}

func (x *XormSnowflake) ToDB() ([]byte, error) {
	return []byte(snowflake.ID(*x).String()), nil
}
