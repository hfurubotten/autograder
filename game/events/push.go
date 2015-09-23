package events

func HandlePush(b []byte) {
	defer PanicHandler(true)
}
