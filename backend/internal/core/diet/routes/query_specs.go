package routes

import "be-simpletracker/internal/utils"

var (
	monthOffsetQuery = utils.QueryIntVar{
		Key:        "monthoffset",
		Default:    0,
		ErrInvalid: "monthoffset must be an integer",
	}
)
