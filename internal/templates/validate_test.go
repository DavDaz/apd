package templates

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		t    Template
		want string
	}{
		{validTemplate(), ""},
		{Template{Name: "n", Description: "d", Sections: []Section{{ID: "s", Title: "S"}}}, "id"},
		{Template{ID: "x", Description: "d", Sections: []Section{{ID: "s", Title: "S"}}}, "name"},
		{Template{ID: "x", Name: "n", Sections: []Section{{ID: "s", Title: "S"}}}, "description"},
		{Template{ID: "x", Name: "n", Description: "d"}, "sections"},
		{Template{ID: "x", Name: "n", Description: "d", Sections: []Section{{Title: "S"}}}, "field id"},
		{Template{ID: "x", Name: "n", Description: "d", Sections: []Section{{ID: "s"}}}, "field title"},
		{Template{ID: "x", Name: "n", Description: "d", Sections: []Section{{ID: "s", Title: "S"}, {ID: "s", Title: "S"}}}, "duplicate section id"},
	}
	for _, tc := range cases {
		err := Validate(tc.t)
		if tc.want == "" && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
			t.Fatalf("error = %v, want substring %q", err, tc.want)
		}
	}
}

func TestLoadRejectsUnknownFieldsAndPreservesOrder(t *testing.T) {
	if _, err := Load(strings.NewReader("id: x\nname: X\ndescription: D\nunknown: y\nsections:\n- id: s\n  title: S\n")); err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("unknown field error = %v", err)
	}
	tmpl, err := Load(strings.NewReader("id: x\nname: X\ndescription: D\nsections:\n- id: first\n  title: First\n- id: second\n  title: Second\n"))
	if err != nil || tmpl.Sections[0].ID != "first" || tmpl.Sections[1].ID != "second" {
		t.Fatalf("Load() = %#v, %v", tmpl, err)
	}
}

func validTemplate() Template {
	return Template{ID: "product", Name: "Product", Description: "Product template", Sections: []Section{{ID: "problem", Title: "Problem"}}}
}
