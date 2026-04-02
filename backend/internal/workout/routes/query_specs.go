package routes

import "be-simpletracker/internal/utils"

var (
	weekOffsetQuery = utils.QueryIntVar{
		Key:        "offset",
		Default:    0,
		ErrInvalid: "offset must be an integer",
	}
	monthOffsetQuery = utils.QueryIntVar{
		Key:        "monthoffset",
		Default:    0,
		ErrInvalid: "monthoffset must be an integer",
	}
)
