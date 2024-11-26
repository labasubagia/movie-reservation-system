package main

import "github.com/jackc/pgx/v5"

func mergeNamedArgs(maps ...pgx.NamedArgs) pgx.NamedArgs {
	merged := pgx.NamedArgs{}
	for _, item := range maps {
		for key, value := range item {
			merged[key] = value
		}
	}
	return merged
}
