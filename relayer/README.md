TODO: make it descriptive

## Why not using goroutines
At the moment, each binary execution expects a private key for broadcasting transactions.
This means we can only use one binary per private key at a time as nonces are autoincremental and must be unique.
That forces us to broadcast transactions sequentially #TX2 needs to be broadcast after #TX1.

If we support more private keys in the future, we may change the code or run two different instances of the binary one for each private key.
