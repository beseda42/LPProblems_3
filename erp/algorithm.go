package erp

// inputData - входные данные задачи.
type inputData struct {
	n int               // число этапов
	k int               // максимальный срок службы оборудования
	r func(age int) int // прибыль в течение одного этапа от использования оборудования, имеющего возраст age
	c func(age int) int // затраты в течение одного этапа на обслуживание оборудования, имеющего возраст age
	s func(age int) int // цена продажи оборудования, имеющего возраст age, в начале этапа
	I int               // цена нового оборудования

}

// Result - результат вычислений.
type Result struct {
	F0               int
	Decisions        []string // оптимальный выбор на каждом этапе
	DecisionsSummary string   // описание оптимальных выборов для задач типа max_test
}

type path struct {
	*inputData

	t int // возраст на начало этапа
	i int // текущий этап

	output []int // текущий результат для каждого возможного t
}

// keep - отдельная прибыль этапа в случае продолжения использования оборудования.
func (inst inputData) keep(age int) int {
	return inst.r(age) - inst.c(age)
}

// sell - отдельная прибыль этапа в случае замены оборудования.
func (inst inputData) sell(age int) int {
	return inst.r(0) - inst.c(0) + inst.s(age) - inst.I
}

// stage - подсчет для всех вариантов возрастов на этапе
func (inst *path) stage() []bool {
	newOutput := make([]int, len(inst.output))
	decision := make([]bool, inst.k+1) // true = keep, false = sell

	for age := 0; age < inst.k+1; age++ {
		sell := inst.sell(age)

		if inst.i == inst.n {
			// на последнем этапе - продажа
			sell += inst.s(1)
		} else {
			// следующее значение
			sell += inst.output[1]
		}

		if age == inst.k { // только продажа
			newOutput[age] = sell
			decision[age] = false
			continue
		}

		keep := inst.keep(age)
		if inst.i == inst.n {
			// на последнем этапе - продажа
			keep += inst.s(age + 1)
		} else {
			// следующее значение
			keep += inst.output[age+1]
		}

		if keep >= sell {
			newOutput[age] = keep
			decision[age] = true
		} else {
			newOutput[age] = sell
			decision[age] = false
		}
	}

	inst.output = newOutput
	return decision
}

// Solve - основная функция решения задачи.
func (inst *path) Solve() Result {
	inst.output = make([]int, inst.k+1)

	// stageDecision[i][age] содержит оптимальный выбор на этапе i для возраста age
	stageDecision := make([][]bool, inst.n+1)
	for i := inst.n; i > 0; i-- {
		inst.i = i
		stageDecision[i] = inst.stage()
	}

	decisions := make([]string, 0, inst.n)
	age := 0
	for stage := 1; stage <= inst.n; stage++ {
		if stageDecision[stage][age] {
			decisions = append(decisions, "keep")
			age++
		} else {
			decisions = append(decisions, "sell")
			age = 1
		}
	}

	return Result{F0: inst.output[0], Decisions: decisions}
}

func (in inputData) solvePeriodic() (int, string) {
	n := in.n
	k := in.k
	I := in.I
	keepProfit := in.r(0) - in.c(0)
	if keepProfit < 0 {
		return -1, "keep profit is negative"
	}
	s := in.s

	var keeps int
	var sells int
	var ageEnd int

	if n <= k {
		keeps = n
		sells = 0
		ageEnd = n
	} else {
		keeps = k
		sells = 1

		// Начинается периодичность
		then := n - (k + 1)
		cycles := then / k // в дальнейшем после sell возраст будет уже 1
		tail := then % k   // остаток

		keeps += cycles*(k-1) + tail
		sells += cycles // по 1 sell на каждый цикл
		ageEnd = tail + 1
	}

	sellAtK := keepProfit + s(k) - I
	stageProfit := keeps*keepProfit + sells*sellAtK
	f0 := stageProfit + s(ageEnd)

	summary := "keep while age<k; when age==k sell (buy new); repeat"
	return f0, summary
}
