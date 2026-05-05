package eka

import (
	"LPProblems/graphs"
	"fmt"
)

func FindMaxFlow(path string) (int, [][]int, error) {
	// Парсим входной файл в структуру graph
	g, err := graphs.ParseDIMACSFile(path)
	if err != nil {
		return 0, nil, err
	}

	g0 := g.Clone()

	maxFlow := 0 // Размер максимального потока
	for {
		// В цикле: ищем кратчайший путь
		path, minVal, ok, err := bfsShortestPath(g)
		if err != nil {
			return 0, nil, err
		}
		if !ok {
			diff, err := graphs.MatrixDifference(g0, g)
			if err != nil {
				return maxFlow, nil, err
			}
			return maxFlow, diff, nil
		}
		for i := 1; i < len(path); i++ {
			// Вычитаем его из графа
			err := g.ChangeCapacity(path[i-1], path[i], minVal)
			if err != nil {
				return 0, nil, err
			}
		}
		maxFlow += minVal
	}
}

// bfsShortestPath ищет путь минимальной длины (в ребрах) от source к sink.
func bfsShortestPath(g *graphs.Graph) ([]int, int, bool, error) {
	// Инициализация
	n := g.VerticesCount()
	source := g.Source()
	sink := g.Sink()
	if source < 0 || source >= n || sink < 0 || sink >= n {
		return nil, 0, false, fmt.Errorf("invalid terminals: source=%d sink=%d", source, sink)
	}

	bestPrev := make([]int, n)
	for i := range bestPrev {
		bestPrev[i] = -1 // по умолчанию - недостижимо
	}

	queue := make([]int, 0, n)
	queue = append(queue, source)

	visited := make([]bool, n)
	visited[source] = true

	// Проходим по всем вершинам - bfs
	for head := 0; head < len(queue) && !visited[sink]; head++ {
		u := queue[head]
		for v := 0; v < n; v++ {
			capacity, ok := g.Capacity(u, v)
			if !ok || capacity <= 0 || visited[v] {
				continue
			}
			visited[v] = true
			bestPrev[v] = u
			queue = append(queue, v)
		}
	}

	// Если путь не найден - завершаем работу
	if !visited[sink] {
		return nil, 0, false, nil
	}

	path, ok := PathFromBestPrev(bestPrev, source, sink)
	if !ok || len(path) < 2 {
		return nil, 0, false, fmt.Errorf("failed to restore bfs path from source=%d to sink=%d", source, sink)
	}

	minVal := 0
	for i := 1; i < len(path); i++ {
		c, has := g.Capacity(path[i-1], path[i])
		if !has || c <= 0 {
			return nil, 0, false, fmt.Errorf("invalid residual edge in path: %d -> %d", path[i-1], path[i])
		}
		if i == 1 || c < minVal {
			minVal = c
		}
	}
	if minVal <= 0 {
		return nil, 0, false, fmt.Errorf("non-positive minimum capacity on path: %d", minVal)
	}

	return path, minVal, true, nil
}

// PathFromBestPrev строит список вершин от source до sink:
// в bestPrev[v] хранится лучший предшественник u для дуги u->v.
// Если по цепочке нельзя дойти до source, ok == false.
func PathFromBestPrev(bestPrev []int, source, sink int) ([]int, bool) {
	n := len(bestPrev)
	if n == 0 || source < 0 || sink < 0 || source >= n || sink >= n {
		return nil, false
	}
	if sink == source {
		return []int{sink}, true
	}

	reversed := make([]int, 0, n)
	seen := make([]bool, n)
	cur := sink
	for {
		// без циклов
		if seen[cur] {
			return nil, false
		}

		seen[cur] = true
		reversed = append(reversed, cur)

		if cur == source {
			// разворачиваем и возвращаем
			path := make([]int, 0, len(reversed))
			for i := len(reversed) - 1; i >= 0; i-- {
				path = append(path, reversed[i])
			}
			return path, true
		}

		// идём дальше
		prev := bestPrev[cur]
		if prev < 0 || prev >= n {
			return nil, false
		}

		cur = prev
	}
}
