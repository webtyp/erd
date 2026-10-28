package erd

import "webtyp.com/router"

const errEncode = "erd: encoding the diagram failed"

// Mount registers a read-only GET at path that answers the diagram's Report (the SVG and
// the findings). The Report is built once, here: models do not change while the app runs.
// It returns the Route so the caller declares the access, which is never optional:
//
//	erd.Mount(r, erd.SchemaPath, d).Requires(erd.ResourceSchema, model.Read)
//
// Never .Public(): the map of tables is a map of what to attack.
func Mount(r router.Router, path string, d Diagram) router.Route {
	rep := d.Report()
	return r.Get(path, func(ctx router.Context) {
		ctx.WriteStatus(200)
		if err := ctx.Encode(&rep); err != nil {
			ctx.WriteStatus(500)
			ctx.Write([]byte(errEncode))
		}
	})
}
