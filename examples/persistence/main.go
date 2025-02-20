package main

import (
	"encoding/json"
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"log"
	"os"
	"path"
	"regexp"
)

type Persister interface {
	Load(map[string]any) error
	State() ([]byte, error)
}

type Storer interface {
	Store(key string, data []byte) error
	Load(key string) ([]byte, error)
}

func WithPersistence(store Storer) func(actor.ReceiveFunc) actor.ReceiveFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			defer next(ctx)
			switch ctx.Message().(type) {
			case actor.Initialized:
				player, ok := ctx.Receiver().(Persister)
				if !ok {
					fmt.Println("Receiver does not implement Persister")
					return
				}
				bytes, err := store.Load(ctx.PID().String())
				if err != nil {
					fmt.Println(err)
					return
				}
				var data map[string]any
				if err = json.Unmarshal(bytes, &data); err != nil {
					log.Fatal(err)
				}
				if err = player.Load(data); err != nil {
					fmt.Println(err)
				}
			case actor.Stopped:
				if player, ok := ctx.Receiver().(Persister); ok {
					bytes, err := player.State()
					if err != nil {
						fmt.Println("couldn't encode player")
						return
					}
					if err = store.Store(ctx.PID().String(), bytes); err != nil {
						fmt.Println("failed to store player")
						return
					}
				}
			}
		}
	}
}

type PlayerJSON struct {
	Username string `json:"username"`
	Health   int    `json:"health"`
}

type Player struct {
	Username string
	Health   int
}

func NewPlayer(username string, health int) actor.Producer {
	return func() actor.Receiver {
		return &Player{
			Username: username,
			Health:   health,
		}
	}
}

func (p *Player) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case TakeDamage:
		p.Health -= msg.Amount
		fmt.Printf(
			"player %s took damage %d, health remaining %d\n",
			p.Username, msg.Amount, p.Health,
		)
	}
}

func (p *Player) Load(data map[string]any) error {
	fmt.Println("loading player:", p.Username)
	p.Username = data["username"].(string)
	p.Health = int(data["health"].(float64))
	return nil
}

func (p *Player) State() ([]byte, error) {
	player := PlayerJSON{
		Username: p.Username,
		Health:   p.Health,
	}
	return json.Marshal(player)
}

type TakeDamage struct {
	Amount int
}

type FileStore struct {
	path string
}

func NewFileStore() *FileStore {
	tmp := "./tmp"
	if err := os.Mkdir(tmp, 0755); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	return &FileStore{path: tmp}
}

func (f *FileStore) Store(key string, data []byte) error {
	key = safeFileName(key)
	return os.WriteFile(path.Join(f.path, key), data, 0755)
}

func (f *FileStore) Load(key string) ([]byte, error) {
	key = safeFileName(key)
	return os.ReadFile(path.Join(f.path, key))
}

func safeFileName(s string) string {
	regex := regexp.MustCompile("[^a-zA-Z0-9]")
	return regex.ReplaceAllString(s, "_")
}

func main() {
	store := NewFileStore()
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}
	pid := engine.Spawn(
		NewPlayer("Doom Guy", 100),
		"player",
		actor.WithID("doom_guy"),
		actor.WithMiddleware(WithPersistence(store)),
	)
	engine.Send(pid, TakeDamage{Amount: 13})
	<-engine.Kill(pid).Done()
}
