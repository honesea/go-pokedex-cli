package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/honesea/pokedexcli/internal/pokeapi"
)

type cliCommand struct {
	name        string
	description string
	command     func(*config, []string) error
}

type config struct {
	client   pokeapi.Client
	next     string
	previous string
	pokedex  map[string]Pokemon
}

type Pokemon struct {
	name       string
	experience int
	height     int
	weight     int
	stats      map[string]int
	types      []string
}

func startRepl() {
	reader := bufio.NewScanner(os.Stdin)

	config := config{
		client:   pokeapi.NewClient(),
		next:     "",
		previous: "",
		pokedex:  make(map[string]Pokemon),
	}

	for {
		fmt.Print("pokedex > ")

		reader.Scan()
		text := cleanInput(reader.Text())

		if len(text) == 0 {
			continue
		}

		commandText := text[0]
		cmd, exists := getCommands()[commandText]
		commandArgs := text[1:]

		if exists {
			err := cmd.command(&config, commandArgs)
			if err != nil {
				fmt.Println("There was an issue proccessing that command")
				fmt.Println("")
			}
		} else {
			fmt.Println("That command doesn't exist")
			fmt.Println("")
		}
	}
}

func cleanInput(input string) []string {
	input = strings.ToLower(input)
	words := strings.Fields(input)
	return words
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the pokedex",
			command:     cmdExit,
		},
		"help": {
			name:        "help",
			description: "List all the available commands",
			command:     cmdHelp,
		},
		"map": {
			name:        "map",
			description: "List the next 20 locations",
			command:     cmdMap,
		},
		"mapb": {
			name:        "mapb",
			description: "List the previous 20 locations",
			command:     cmdMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore all the pokemon in a given location",
			command:     cmdExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch pokemon with the given name",
			command:     cmdCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a pokemon with the given name",
			command:     cmdInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all of the pokemon you've caught",
			command:     cmdPokedex,
		},
	}
}

func cmdExit(c *config, args []string) error {
	fmt.Println("Thanks for using the pokedex")
	fmt.Println("")
	os.Exit(0)
	return nil
}

func cmdHelp(c *config, args []string) error {
	fmt.Println("Welcome to the pokedex!")
	fmt.Println("")
	fmt.Println("commands:")

	for _, cmd := range getCommands() {
		fmt.Printf("%s - %s\n", cmd.name, cmd.description)
	}

	fmt.Println("")
	return nil
}

func cmdMap(c *config, args []string) error {
	locations, err := pokeapi.GetLocations(&c.client, &c.next)
	if err != nil {
		fmt.Println("Couldn't list locations")
		return nil
	}

	if locations.Next != nil {
		c.next = *locations.Next
	}

	if locations.Previous != nil {
		c.previous = *locations.Previous
	}

	for i := range locations.Results {
		fmt.Println(locations.Results[i].Name)
	}

	return nil
}

func cmdMapb(c *config, args []string) error {
	if c.previous == "" {
		fmt.Println("No locations to navigate back to")
		return errors.New("no locations to navigate back to")
	}

	locations, err := pokeapi.GetLocations(&c.client, &c.previous)
	if err != nil {
		fmt.Println("Couldn't list locations")
		return nil
	}

	if locations.Next != nil {
		c.next = *locations.Next
	}

	if locations.Previous != nil {
		c.previous = *locations.Previous
	}

	for i := range locations.Results {
		fmt.Println(locations.Results[i].Name)
	}

	return nil
}

func cmdExplore(c *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("You must specify an area to explore")
		return nil
	}

	location, err := pokeapi.GetLocation(&c.client, args[0])
	if err != nil {
		fmt.Println("Couldn't explore that area")
		return nil
	}

	fmt.Printf("Exploring %v\n", args[0])
	for i := range location.PokemonEncounters {
		fmt.Printf(" - %v\n", location.PokemonEncounters[i].Pokemon.Name)
	}

	return nil
}

func cmdCatch(c *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("You must specify a pokemon to catch")
		return nil
	}

	if _, ok := c.pokedex[args[0]]; ok {
		fmt.Println("You have already caught this pokemon")
		return nil
	}

	pokemon, err := pokeapi.GetPokemon(&c.client, args[0])
	if err != nil {
		fmt.Println("Couldn't catch the pokemon")
		return nil
	}

	fmt.Printf("Catching %v ...\n", args[0])
	if randInt := rand.Intn(350); pokemon.BaseExperience > 0 && randInt > pokemon.BaseExperience {
		// Parse stats
		stats := make(map[string]int)
		for _, stat := range pokemon.Stats {
			stats[stat.Stat.Name] = stat.BaseStat
		}

		// Parse types
		types := make([]string, 0)
		for _, pokemonType := range pokemon.Types {
			types = append(types, pokemonType.Type.Name)
		}

		c.pokedex[args[0]] = Pokemon{
			name:       args[0],
			experience: pokemon.BaseExperience,
			height:     pokemon.Height,
			weight:     pokemon.Weight,
			stats:      stats,
			types:      types,
		}
		fmt.Printf("You caught the %v!\n", args[0])
	} else {
		fmt.Println("You failed to catch the pokemon")
	}

	return nil
}

func cmdInspect(c *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("You must specify a pokemon to inspect")
		return nil
	}

	pokemon, ok := c.pokedex[args[0]]
	if !ok {
		fmt.Println("You haven't caught this pokemon")
		return nil
	}

	fmt.Printf("Name: %v\n", pokemon.name)
	fmt.Printf("Height: %v\n", pokemon.height)
	fmt.Printf("Weight: %v\n", pokemon.weight)

	fmt.Println("Stats:")
	for name, value := range pokemon.stats {
		fmt.Printf("  -%v: %v\n", name, value)
	}

	fmt.Println("Types:")
	for _, value := range pokemon.types {
		fmt.Printf("  - %v\n", value)
	}

	return nil
}

func cmdPokedex(c *config, args []string) error {
	if len(c.pokedex) == 0 {
		fmt.Println("You haven't caught in pokemon yet")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for _, pokemon := range c.pokedex {
		fmt.Printf("  -%v\n", pokemon.name)
	}

	return nil
}
