package ast

// NOTE: technically Type- and Value- dependencies are not accurate names since it's more like
// "X needs to be computed before Y" or "X can be computed at the same time as Y",
// regardless of whether it is the type or the value of X and Y
type DependencyGraph struct {
	// Dependencies that this node requires to be able to calculate its type.
	TypeDependencies map[Node][]IName
	// Dependencies that this node requires to be able to calculate its value.
	// Superset of TypeDependencies, though they do not need to be manually added to both.
	ValueDependencies map[Node][]IName
}

func (g DependencyGraph) AddTypeDependency(n Node, dependency IName) {
	existingDependencies, _ := g.TypeDependencies[n]
	g.TypeDependencies[n] = append(existingDependencies, dependency)
}

func (g DependencyGraph) AddTypeDependencies(n Node, dependencies []IName) {
	existingDependencies, _ := g.TypeDependencies[n]
	g.TypeDependencies[n] = append(existingDependencies, dependencies...)
}

func (g DependencyGraph) AddValueDependency(n Node, dependency IName) {
	existingDependencies, _ := g.ValueDependencies[n]
	g.ValueDependencies[n] = append(existingDependencies, dependency)
}

func (g DependencyGraph) AddValueDependencies(n Node, dependencies []IName) {
	existingDependencies, _ := g.ValueDependencies[n]
	g.ValueDependencies[n] = append(existingDependencies, dependencies...)
}

func (g DependencyGraph) AddTypeDependenciesOf(n Node, of Node) {
	g.AddTypeDependencies(n, g.TypeDependencies[of])
}

func (g DependencyGraph) AddValueDependenciesOf(n Node, of Node) {
	g.AddTypeDependenciesOf(n, of)
	g.AddValueDependencies(n, g.ValueDependencies[of])
}
