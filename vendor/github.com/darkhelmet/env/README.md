# env

Easy environment variables for Go

# Usage

    import "github.com/darkhelmet/env"

    ...

    s := env.String("USER") // Will panic if USER is not present
    sd := env.StringDefault("KEY", "It-Not-Present")
    sdf := env.StringDefaultF("KEY", func() string { return "do something tough" })

    // Similarly for int and float

    env.Int("N") // Panic if not present
    env.IntDefault("N", 1)
    env.IntDefaultF("N", func() int { return 5 })

    env.Float("F") // Panic if not present
    env.FloatDefault("F", 1.0)
    env.FloatDefaultF("F", func() float { return 5.5 })

# License

Apache 2.0, see LICENSE.md
