package graphs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ParseDIMACSFile читает файл DIMACS max-flow по пути path.
func ParseDIMACSFile(path string) (*Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()
	return ParseDIMACS(f)
}

// ParseDIMACS разбирает сеть DIMACS max-flow из reader.
func ParseDIMACS(reader io.Reader) (*Graph, error) {
	scanner := bufio.NewScanner(reader) // scanner читает вход построчно.
	lineN := 0                          // lineN — номер текущей строки во входе.

	var graph *Graph    // graph — итоговый граф, создается после строки p max.
	declaredArcs := -1  // declaredArcs — число дуг из problem line.
	parsedArcs := 0     // parsedArcs — сколько дуг реально прочитано.
	sourceSeen := false // sourceSeen — встретился ли дескриптор истока.
	sinkSeen := false   // sinkSeen — встретился ли дескриптор стока.
	seenArc := false    // seenArc — началась ли уже секция дуг.

	for scanner.Scan() {
		lineN++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "c":
			continue
		case "p":
			if graph != nil {
				return nil, fmt.Errorf("line %d: multiple problem lines", lineN)
			}
			if len(fields) < 4 {
				return nil, fmt.Errorf("line %d: invalid problem line", lineN)
			}
			if fields[1] != "max" {
				return nil, fmt.Errorf("line %d: unsupported problem type %q", lineN, fields[1])
			}
			n, err := strconv.Atoi(fields[2]) // n — количество вершин.
			if err != nil {
				return nil, fmt.Errorf("line %d: parse node count: %w", lineN, err)
			}
			m, err := strconv.Atoi(fields[3]) // m — количество дуг из заголовка.
			if err != nil {
				return nil, fmt.Errorf("line %d: parse arc count: %w", lineN, err)
			}
			if m < 0 {
				return nil, fmt.Errorf("line %d: arc count must be non-negative, got %d", lineN, m)
			}
			graph, err = NewGraph(n)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineN, err)
			}
			declaredArcs = m
		case "n":
			if graph == nil {
				return nil, fmt.Errorf("line %d: node descriptor before problem line", lineN)
			}
			if seenArc {
				return nil, fmt.Errorf("line %d: node descriptor after arc descriptors", lineN)
			}
			if len(fields) < 3 {
				return nil, fmt.Errorf("line %d: invalid node descriptor", lineN)
			}
			id, err := strconv.Atoi(fields[1]) // id — 1-based id вершины из файла.
			if err != nil {
				return nil, fmt.Errorf("line %d: parse node id: %w", lineN, err)
			}
			v := id - 1 // v — внутренний 0-based id вершины.
			if err := graph.validateVertex(v); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineN, err)
			}
			switch fields[2] {
			case "s":
				if sourceSeen {
					return nil, fmt.Errorf("line %d: duplicate source descriptor", lineN)
				}
				graph.source = v
				sourceSeen = true
			case "t":
				if sinkSeen {
					return nil, fmt.Errorf("line %d: duplicate sink descriptor", lineN)
				}
				graph.sink = v
				sinkSeen = true
			default:
				return nil, fmt.Errorf("line %d: unsupported node type %q", lineN, fields[2])
			}
		case "a":
			if graph == nil {
				return nil, fmt.Errorf("line %d: arc descriptor before problem line", lineN)
			}
			if !sourceSeen || !sinkSeen {
				return nil, fmt.Errorf("line %d: arc descriptor before source/sink descriptors", lineN)
			}
			if len(fields) < 4 {
				return nil, fmt.Errorf("line %d: invalid arc descriptor", lineN)
			}
			src, err := strconv.Atoi(fields[1]) // src — 1-based начало дуги из файла.
			if err != nil {
				return nil, fmt.Errorf("line %d: parse arc source: %w", lineN, err)
			}
			dst, err := strconv.Atoi(fields[2]) // dst — 1-based конец дуги из файла.
			if err != nil {
				return nil, fmt.Errorf("line %d: parse arc destination: %w", lineN, err)
			}
			cap64, err := strconv.ParseInt(fields[3], 10, 32) // cap64 — capacity основной дуги.
			if err != nil {
				return nil, fmt.Errorf("line %d: parse arc capacity: %w", lineN, err)
			}
			if err := addSignedArcNoConflict(graph, src-1, dst-1, int(cap64)); err != nil {
				return nil, fmt.Errorf("line %d: add arc: %w", lineN, err)
			}
			parsedArcs++

			// В части файлов используется форма "a u v cap_uv cap_vu".
			// В этом случае добавляется и обратная ориентированная дуга.
			if len(fields) >= 5 {
				revCap64, err := strconv.ParseInt(fields[4], 10, 32) // revCap64 — capacity обратной дуги.
				if err != nil {
					return nil, fmt.Errorf("line %d: parse reverse arc capacity: %w", lineN, err)
				}
				if err := addSignedArcNoConflict(graph, dst-1, src-1, int(revCap64)); err != nil {
					return nil, fmt.Errorf("line %d: add reverse arc: %w", lineN, err)
				}
				parsedArcs++
			}
			seenArc = true
		default:
			return nil, fmt.Errorf("line %d: unsupported directive %q", lineN, fields[0])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}
	if graph == nil {
		return nil, fmt.Errorf("missing problem line")
	}
	if !sourceSeen || !sinkSeen {
		return nil, fmt.Errorf("missing source or sink node descriptor")
	}
	// В некоторых реальных данных число дуг в заголовке может не совпадать.
	// Проверяем только то, что разобрано не меньше заявленного.
	if declaredArcs >= 0 && parsedArcs < declaredArcs {
		return nil, fmt.Errorf("declared at least %d arcs but parsed %d", declaredArcs, parsedArcs)
	}
	if err := graph.Validate(); err != nil {
		return nil, err
	}
	return graph, nil
}

func addArcNoConflict(graph *Graph, src, dst int, cap int) error {
	if cap <= 0 {
		return nil // нулевая пропускная способность — дуги нет, строку игнорируем
	}
	if existingCap, ok := graph.Capacity(src, dst); ok {
		if existingCap != cap {
			return fmt.Errorf("conflicting capacities for arc (%d, %d): %d vs %d", src, dst, existingCap, cap)
		}
		return nil
	}
	return graph.AddArc(src, dst, cap)
}

// addSignedArcNoConflict трактует отрицательную capacity как дугу в обратную сторону.
func addSignedArcNoConflict(graph *Graph, src, dst int, cap int) error {
	if cap < 0 {
		src, dst = dst, src
		cap = -cap
	}
	return addArcNoConflict(graph, src, dst, cap)
}
