package tests

import "webtyp.com/model"

// fakeModel is a table as ormc-generated models present it: a schema plus SchemaExt for the foreign keys.
type fakeModel struct {
	name   string
	fields []model.Field
	exts   []model.FieldExt
}

func (m *fakeModel) ModelName() string                { return m.name }
func (m *fakeModel) Schema() []model.Field            { return m.fields }
func (m *fakeModel) Pointers() []any                  { return nil }
func (m *fakeModel) IsNil() bool                      { return m == nil }
func (m *fakeModel) EncodeFields(_ model.FieldWriter) {}
func (m *fakeModel) DecodeFields(_ model.FieldReader) {}
func (m *fakeModel) SchemaExt() []model.FieldExt      { return m.exts }

func newTable(name string, cols ...model.Field) *fakeModel {
	return &fakeModel{name: name, fields: cols}
}

// withFK declares column as a foreign key to ref.
func (m *fakeModel) withFK(column, ref string) *fakeModel {
	for _, f := range m.fields {
		if f.Name == column {
			m.exts = append(m.exts, model.FieldExt{Field: f, Ref: ref})
		}
	}
	return m
}

func pk(name string, k model.Kind) model.Field {
	return model.Field{Name: name, Type: k, NotNull: true, DB: &model.FieldDB{PK: true}}
}

func col(name string, k model.Kind) model.Field {
	return model.Field{Name: name, Type: k}
}

func tenant() *fakeModel {
	return newTable("tenant", pk("id", model.Text()), col("name", model.Text()))
}

func user() *fakeModel {
	return newTable("user", pk("id", model.Text()), col("tenant_id", model.Text())).withFK("tenant_id", "tenant")
}
