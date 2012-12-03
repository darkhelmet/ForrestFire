package retry

func Times(times int, f func() error) {
    for i := 0; i < times; i++ {
        err := f()
        if err == nil {
            return
        }
    }
}
