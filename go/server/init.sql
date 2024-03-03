CREATE TABLE IF NOT EXISTS "witness"(
                                        ID TEXT PRIMARY KEY ,
                                        witness_level TEXT,
                                        witness_dep TEXT);

CREATE TABLE IF NOT EXISTS "accumulator"(
                                            acc_type TEXT,
                                            acc_value TEXT,
                                            acc TEXT
);

CREATE TABLE IF NOT EXISTS "auth"(
                                     id TEXT,
                                     level TEXT,
                                     dep TEXT);

CREATE TABLE IF NOT EXISTS "user"(
                                     id PRIMARY KEY ,
                                     pk TEXT,
);