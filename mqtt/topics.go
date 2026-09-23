package mqtt

type Topic struct {
	Name string
	QoS  byte
}

var Topics = []Topic{
	{
		Name: "my/test/topic",
		QoS:  1,
	},
}
