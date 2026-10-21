package mqtt

type Topic struct {
	Name string
	QoS  byte
}

var Topics = []Topic{
	{
		Name: "/check_role",
		QoS:  1,
	},
	{
		Name: "/pay",
		QoS:  1,
	},
	{
		Name: "/retieve_member",
		QoS:  1,
	},
	{
		Name: "/retieve_leaderboard",
		QoS:  1,
	},
	{
		Name: "/add_money",
		QoS:  1,
	},
	{
		Name: "/add_score",
		QoS:  1,
	},
	{
		Name: "/create_guest",
		QoS:  1,
	},
	{
		Name: "/create_member",
		QoS:  1,
	},
	{
		Name: "/remove_card",
		QoS:  1,
	},
}
