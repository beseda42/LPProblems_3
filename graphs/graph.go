package graphs

import (
	"fmt"
)

// Graph - ориентированная сеть с истоком source и стоком sink.
// capacity[u][v] > 0 означает дугу u -> v с этой пропускной способностью; 0 — дуги нет.
type Graph struct {
	N      int // количество вершин в графе
	source int // индекс истока
	sink   int // индекс стока

	capacity [][]int // capacity[u][v] — пропускная способность дуги u -> v (только >= 0)
	arcCount int     // число дуг с положительной capacity
}

// IncomingArc — входящая дуга u -> заданная вершина с пропускной способностью.
type IncomingArc struct {
	From     int // вершина-источник дуги
	Capacity int // capacity дуги From -> v
}

// NewGraph создает пустой граф из n вершин.
func NewGraph(n int) (*Graph, error) {
	if n <= 0 {
		return nil, fmt.Errorf("graph must have positive number of vertices, got %d", n)
	}
	capacity := make([][]int, n)
	for i := 0; i < n; i++ {
		capacity[i] = make([]int, n)
	}
	return &Graph{
		N:        n,
		source:   -1,
		sink:     -1,
		capacity: capacity,
	}, nil
}

// Clone возвращает копию графа.
func (g *Graph) Clone() *Graph {
	if g == nil {
		return nil
	}
	h := &Graph{
		N:        g.N,
		source:   g.source,
		sink:     g.sink,
		arcCount: g.arcCount,
		capacity: make([][]int, g.N),
	}
	for i := 0; i < g.N; i++ {
		h.capacity[i] = make([]int, g.N)
		copy(h.capacity[i], g.capacity[i])
	}
	return h
}

// MatrixDifference возвращает матрицу d, где d[i][j] = g0[i][j] - g1[i][j]
// (поэлементная разность пропускных способностей). Размеры g0 и g1 должны совпадать.
func MatrixDifference(g0, g1 *Graph) ([][]int, error) {
	if g0 == nil || g1 == nil {
		return nil, fmt.Errorf("nil graph")
	}
	if g0.N != g1.N {
		return nil, fmt.Errorf("graph sizes differ: %d vs %d", g0.N, g1.N)
	}
	d := make([][]int, g0.N)
	for i := 0; i < g0.N; i++ {
		d[i] = make([]int, g0.N)
		for j := 0; j < g0.N; j++ {
			d[i][j] = g0.capacity[i][j] - g1.capacity[i][j]
		}
	}
	return d, nil
}

// SetTerminals задает вершины истока source и стока sink.
func (g *Graph) SetTerminals(source, sink int) error {
	if err := g.validateVertex(source); err != nil {
		return err
	}
	if err := g.validateVertex(sink); err != nil {
		return err
	}
	if source == sink {
		return fmt.Errorf("source and sink must be different: %d", source)
	}
	g.source = source
	g.sink = sink
	return nil
}

// AddArc добавляет ориентированную дугу (src -> dst) с capacity cap.
func (g *Graph) AddArc(src, dst int, cap int) error {
	if err := g.validateVertex(src); err != nil {
		return err
	}
	if err := g.validateVertex(dst); err != nil {
		return err
	}
	if src == dst {
		return fmt.Errorf("self loops are not allowed: %d", src)
	}
	if cap <= 0 {
		return fmt.Errorf("capacity must be positive, got %d", cap)
	}
	if g.capacity[src][dst] > 0 {
		return fmt.Errorf("parallel arcs are not allowed: (%d, %d)", src, dst)
	}
	g.capacity[src][dst] = cap
	g.arcCount++
	return nil
}

// HasArc возвращает true, если дуга (src -> dst) существует.
func (g *Graph) HasArc(src, dst int) bool {
	if src < 0 || dst < 0 || src >= g.N || dst >= g.N {
		return false
	}
	return g.capacity[src][dst] > 0
}

// Capacity возвращает capacity дуги (src -> dst).
// Второе возвращаемое значение показывает, существует ли дуга.
func (g *Graph) Capacity(src, dst int) (int, bool) {
	if src < 0 || dst < 0 || src >= g.N || dst >= g.N {
		return 0, false
	}
	c := g.capacity[src][dst]
	if c <= 0 {
		return 0, false
	}
	return c, true
}

// ChangeCapacity уменьшает capacity дуги u -> v на число value и добавляет value к capacity дуги v -> u.
func (g *Graph) ChangeCapacity(u, v, value int) error {
	if err := g.validateVertex(u); err != nil {
		return err
	}
	if err := g.validateVertex(v); err != nil {
		return err
	}
	if u == v {
		return fmt.Errorf("self loops are not allowed: %d", u)
	}
	if value < 0 {
		return fmt.Errorf("value must be non-negative, got %d", value)
	}
	if value == 0 {
		return nil
	}
	if g.capacity[u][v] <= 0 {
		return fmt.Errorf("no arc (%d -> %d)", u, v)
	}
	if g.capacity[u][v] < value {
		return fmt.Errorf("insufficient capacity on (%d -> %d): have %d, need %d", u, v, g.capacity[u][v], value)
	}
	g.capacity[u][v] -= value
	if g.capacity[u][v] == 0 {
		g.arcCount--
	}
	if g.capacity[v][u] > 0 {
		g.capacity[v][u] += value
	} else {
		g.capacity[v][u] = value
		g.arcCount++
	}
	return nil
}

// ArcCount возвращает число добавленных ориентированных дуг.
func (g *Graph) ArcCount() int {
	return g.arcCount
}

// VerticesCount возвращает число вершин в графе
func (g *Graph) VerticesCount() int {
	return g.N
}

func (g *Graph) Source() int {
	return g.source
}
func (g *Graph) Sink() int {
	return g.sink
}

// Validate проверяет корректность внутренней структуры графа.
func (g *Graph) Validate() error {
	if g.N <= 0 {
		return fmt.Errorf("graph must have positive number of vertices")
	}
	if len(g.capacity) != g.N {
		return fmt.Errorf("adjacency matrix dimensions do not match N=%d", g.N)
	}
	count := 0
	for i := 0; i < g.N; i++ {
		if len(g.capacity[i]) != g.N {
			return fmt.Errorf("matrix row %d has invalid length", i)
		}
		if g.capacity[i][i] != 0 {
			return fmt.Errorf("self loop at vertex %d", i)
		}
		for j := 0; j < g.N; j++ {
			c := g.capacity[i][j]
			if c < 0 {
				return fmt.Errorf("negative capacity at (%d, %d): %d", i, j, c)
			}
			if c > 0 {
				count++
			}
		}
	}
	if count != g.arcCount {
		return fmt.Errorf("internal arc count mismatch: have %d, counted %d", g.arcCount, count)
	}
	if err := g.validateTerminals(); err != nil {
		return err
	}
	return nil
}

func (g *Graph) validateVertex(v int) error {
	if v < 0 || v >= g.N {
		return fmt.Errorf("vertex %d is out of range [0, %d)", v, g.N)
	}
	return nil
}

func (g *Graph) validateTerminals() error {
	if g.source < 0 || g.source >= g.N {
		return fmt.Errorf("invalid source id: %d", g.source)
	}
	if g.sink < 0 || g.sink >= g.N {
		return fmt.Errorf("invalid sink id: %d", g.sink)
	}
	if g.source == g.sink {
		return fmt.Errorf("source and sink must be distinct")
	}
	return nil
}
