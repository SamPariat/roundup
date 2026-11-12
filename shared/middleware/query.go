package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// locationKeys are query parameters holding where a user is. They are logged
// rounded to locationPlaces decimal places, about 1.1 km, which is enough to
// debug "no results here" without logs keeping anyone's exact location.
var locationKeys = map[string]struct{}{
	"lat": {},
	"lng": {},
}

// locationPlaces is how many decimal places a logged location keeps.
const locationPlaces = 2

// queryField logs the request's query string as a "query" object of its
// parameters (the last value wins for a repeated key), with locations rounded
// and sensitive keys redacted (see loggedQueryValue). No query string adds
// nothing.
func queryField(c fiber.Ctx) zap.Field {
	params := c.Queries()
	if len(params) == 0 {
		return zap.Skip()
	}

	logged := make(map[string]string, len(params))
	for key, value := range params {
		logged[key] = loggedQueryValue(key, value)
	}

	return zap.Any("query", logged)
}

// loggedQueryValue is what the logs keep of one query parameter: a sensitive
// key's value redacted, a location rounded (or redacted when it isn't a
// number, since it could then hold anything), and any other value whole.
func loggedQueryValue(key, value string) string {
	if _, ok := sensitiveKeys[key]; ok {
		return redacted
	}

	if _, ok := locationKeys[key]; ok {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return redacted
		}

		return strconv.FormatFloat(f, 'f', locationPlaces, 64)
	}

	return value
}
