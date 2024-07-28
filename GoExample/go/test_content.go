package main

func Content() {
	var j = 10
	if j > 10 {
		return
	}
	var i = 13
	for i = 0; i < 10; i++ {
		var s string
		if s != "" {
			switch s {
			case "1":
				return
			case "2":
				if i > 10 {
					break
				}
			default:
				for {
					var s chan int
					select {
					case <-s:
					default:
						for ; i > 0; i++ {

						}
					}
				}
			}
		}
	}
}
