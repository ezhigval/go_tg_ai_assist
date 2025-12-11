package bot

import "sync"

// State представляет состояние пользователя в FSM.
// Name — имя состояния (например, "TODO_ADD", "FIN_ADD_TYPE").
// Data — произвольные данные состояния (map для гибкости).
type State struct {
	Name string
	Data map[string]any
}

// FSM (Finite State Machine) управляет состояниями пользователей в памяти.
// Используется для многошаговых диалогов (создание задач, финансовых операций и т.д.).
// Состояния хранятся в памяти и очищаются после завершения диалога.
type FSM struct {
	mu    sync.RWMutex
	store map[int64]*State
}

// NewFSM создаёт новый экземпляр FSM.
func NewFSM() *FSM {
	return &FSM{
		store: make(map[int64]*State),
	}
}

// Set устанавливает состояние для пользователя.
// Если состояние уже существует, оно перезаписывается.
func (f *FSM) Set(userID int64, name string, data map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[userID] = &State{Name: name, Data: data}
}

// Get возвращает текущее состояние пользователя или nil, если состояние не установлено.
func (f *FSM) Get(userID int64) *State {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.store[userID]
}

// Clear удаляет состояние пользователя (завершение диалога).
func (f *FSM) Clear(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.store, userID)
}
