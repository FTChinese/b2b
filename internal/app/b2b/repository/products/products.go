package products

import (
	"github.com/FTChinese/ftacademy/internal/app/b2b/stmt"
	"github.com/FTChinese/ftacademy/internal/pkg/plan"
)

func (env Env) LoadProducts() ([]plan.Product, error) {
	// TODO: find in cache.

	productRows, err := env.retrieveProducts()
	if err != nil {
		return nil, err
	}

	planIDs := plan.GetProductsPlanIDs(productRows)

	groupedPlans, err := env.PlansInSet(planIDs)
	if err != nil {
		return nil, err
	}

	products := plan.ZipProductWithPlan(productRows, groupedPlans)

	// TODO: cache the final products.

	return products, nil
}

func (env Env) retrieveProducts() ([]plan.ProductSchema, error) {
	var rows []plan.ProductSchema
	if err := env.db.Select(&rows, stmt.Products); err != nil {
		return nil, err
	}

	return rows, nil
}
