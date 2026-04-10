package routes

import "be-simpletracker/internal/utils"

var (
	monthOffsetQuery = utils.QueryIntVar{
		Key:        "monthoffset",
		Default:    0,
		ErrInvalid: "monthoffset must be an integer",
	}
	pageQuery = utils.QueryIntVar{
		Key:        "page",
		Default:    1,
		ErrInvalid: "page must be an integer",
	}
	pageSizeQuery = utils.QueryIntVar{
		Key:        "page_size",
		Default:    0,
		ErrInvalid: "page_size must be an integer",
	}
	activityWeeksQuery = utils.QueryIntVar{
		Key:        "weeks",
		Default:    52,
		ErrInvalid: "weeks must be an integer",
	}
)
