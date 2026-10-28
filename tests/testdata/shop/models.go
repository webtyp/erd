package shop

import "webtyp.com/model"

var CustomerModel = model.Definition{
	Name: "customer",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: model.Text()},
	},
}

var InvoiceModel = model.Definition{
	Name: "invoice",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "customer_id", Type: model.Text(), Ref: &CustomerModel},
	},
}
