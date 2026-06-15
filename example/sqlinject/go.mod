module sqlinjectdemo

go 1.22

replace sqlinjecthook => ./rules

require github.com/go-sql-driver/mysql v1.8.1

require filippo.io/edwards25519 v1.1.0 // indirect
