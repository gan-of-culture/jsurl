package jsurl

import "testing"

type demoSubStruct struct {
	D string
	E demoStruct
	F []int
}

type demoMixStruct struct {
	A interface{}
	B interface{}
	C interface{}
}

type demoStruct struct {
	A func()
	B interface{}
	C bool
	D int
	E string
}

func TestBasicValues(t *testing.T) {

	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{
			name: "function",
			in:   func() {},
			want: "",
		}, {
			name: "nil",
			in:   nil,
			want: "~null",
		}, {
			name: "false",
			in:   false,
			want: "~false",
		}, {
			name: "true",
			in:   true,
			want: "~true",
		}, {
			name: "0",
			in:   0,
			want: "~0",
		}, {
			name: "1",
			in:   1,
			want: "~1",
		}, {
			name: "negativ float",
			in:   -1.5,
			want: "~-1.5",
		}, {
			name: "unicode char",
			in:   "hello world\u203c",
			want: "~'hello*20world**203c",
		}, {
			name: "special characters",
			in:   " !\"#$%&'()*+,-./09:;<=>?@AZ[\\]^_`az{|}~",
			want: "~'*20*21*22*23!*25*26*27*28*29*2a*2b*2c-.*2f09*3a*3b*3c*3d*3e*3f*40AZ*5b*5c*5d*5e_*60az*7b*7c*7d*7e",
		}, {
			name: "slice empty",
			in:   []string{},
			want: "~(~)",
		}, {
			name: "slice",
			in:   []interface{}{nil, func() {}, false, 0, "hello world\u203c"},
			want: "~(~null~null~false~0~'hello*20world**203c)",
		}, {
			name: "struct empty",
			in: struct {
				A string
				B string
			}{},
			want: "~()",
		}, {
			name: "struct",
			in: demoStruct{
				A: func() {},
				B: nil,
				C: false,
				D: 0,
				E: "hello world\u203c",
			},
			want: "~(B~null~C~false~D~0~E~'hello*20world**203c)",
		}, {
			name: "mix",
			in: demoMixStruct{
				A: []interface{}{
					[]int{1, 2},
					[]int{},
					demoStruct{},
				},
				B: []int{},
				C: demoSubStruct{
					D: "hello",
					E: demoStruct{},
					F: []int{}},
			},
			want: "~(A~(~(~1~2)~(~)~())~B~(~)~C~(D~'hello~E~()~F~(~)))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := Stringify(tt.in)
			if val != tt.want {
				t.Errorf("Got: %v - want: %v", val, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type want struct {
		A string
		B string
	}

	tests := []struct {
		name string
		in   string
		want want
	}{
		{
			name: "percent-escaped single quotes",
			in:   `~(a~%27hello~b~%27world)`,
			want: want{
				A: "hello",
				B: "world",
			},
		}, {
			name: "percent-escaped percent-escaped single quotes",
			in:   `~(a~%2527hello~b~%2525252527world)`,
			want: want{
				A: "hello",
				B: "world",
			},
		}, {
			name: "dollar sign $$",
			in:   `~(a~%2527hello~b~%2525252527world!)`,
			want: want{
				A: "hello",
				B: "world$",
			},
		}, {
			name: "dollar sign $$",
			in:   `~(a~%2527hello*20~b~%2525252527world!)`,
			want: want{
				A: "hello ",
				B: "world$",
			},
		}, {
			name: "unicode char",
			in:   `~(a~%2527hello*20~b~%2525252527world!**203c)`,
			want: want{
				A: "hello ",
				B: "world$\u203c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := want{}
			Parse(tt.in, &val)
			if val != tt.want {
				t.Errorf("Got: %v - want: %v", val, tt.want)
			}
		})
	}
}

type tag struct {
	BooksCount  uint   `json:"books_count"`
	Description string `json:"description"`
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Type        uint   `json:"type"`
}

type items struct {
	Excluded []tag `json:"excluded"`
	Included []tag `json:"included"`
}

type searchTag struct {
	Items items         `json:"items"`
	Tags  []interface{} `json:"tags"`
	Text  string        `json:"text"`
	Type  uint          `json:"type"`
}

type pages struct {
	Range []uint `json:"range"`
}

type search struct {
	Page  uint      `json:"page"`
	Pages pages     `json:"pages"`
	Sort  uint      `json:"sort"`
	Tag   searchTag `json:"tag"`
	Text  string    `json:"text"`
}

func TestParseDeepStruct(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want search
	}{
		{
			name: "sample search query",
			in:   `~(text~'~page~0~sort~0~pages~(range~(~0~2000))~tag~(text~'F~type~1~tags~(~)~items~(included~(~(id~1101~name~'AG~description~null~type~1~books_count~48)~(id~232~name~'FRY~description~null~type~1))~excluded~(~))))`,
			want: search{
				Text: "",
				Page: 0,
				Sort: 4,
				Pages: pages{
					Range: []uint{0, 2000},
				},
				Tag: searchTag{
					Text: "",
					Items: items{
						Included: []tag{
							{
								BooksCount:  48,
								Description: "",
								ID:          1101,
								Name:        "AG",
								Type:        1,
							}, {
								Description: "",
								ID:          232,
								Name:        "FRY",
								Type:        1,
							},
						},
						Excluded: []tag{},
					},
					Tags: []interface{}{},
					Type: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := search{}
			Parse(tt.in, &val)
			if val.Tag.Items.Included[0] != tt.want.Tag.Items.Included[0] || val.Tag.Items.Included[1] != tt.want.Tag.Items.Included[1] {
				t.Errorf("Got: %v - want: %v", val, tt.want)
			}
		})
	}
}
