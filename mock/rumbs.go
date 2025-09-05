package mock

type Rumb struct { // вот наша новая структура
	ID    int    // поля структур, которые передаются в шаблон
	Title string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
	Image string
}

var Rumbs = []Rumb{ // массив элементов из наших структур
	{
		ID:    1,
		Title: "Северо-Запад",
		Image: "arrow-north-west.svg",
	},
	{
		ID:    2,
		Title: "Север",
		Image: "arrow-north.svg",
	},
	{
		ID:    3,
		Title: "Северо-Восток",
		Image: "arrow-north-east.svg",
	},
	{
		ID:    4,
		Title: "Запад",
		Image: "arrow-west.svg",
	},
	{
		ID:    5,
		Title: "Восток",
		Image: "arrow-east.svg",
	},
	{
		ID:    6,
		Title: "Юго-Запад",
		Image: "arrow-south-west.svg",
	},
	{
		ID:    7,
		Title: "Юг",
		Image: "arrow-south.svg",
	},
	{
		ID:    8,
		Title: "Юго-Восток",
		Image: "arrow-south-east.svg",
	},
}
