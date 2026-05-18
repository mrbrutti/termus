// sp16-event-count reports the number of generated NoteEvents per role
// per section for a .tm file. Used to demonstrate the intent-driven
// engine's leverage: author writes N events, engine produces M events.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/track"
)

func main() {
	flag.Parse()
	for _, path := range flag.Args() {
		bytes, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			continue
		}
		file, err := track.Parse(bytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			continue
		}
		fmt.Printf("=== %s ===\n", path)
		compiled, err := track.Compile(file, file.Seed, gen.ListeningModeEndless)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  compile failed: %v\n", err)
			continue
		}
		for key, plan := range compiled.Plans {
			fmt.Printf("  section %s\n", key)
			for _, t := range plan.Tracks {
				count := 0
				for _, n := range t.Notes {
					if n >= 0 {
						count++
					}
				}
				fmt.Printf("    %-15s  %4d notes  channel=%d program=%d reverb=%d\n", t.Name, count, t.Channel, t.Program, t.Reverb)
			}
		}
		fmt.Println()
	}
}
