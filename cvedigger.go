package woodpecker

import "fmt"

type DigOpts struct {
	KillOpts
}

func (w Woodpecker) Dig(args []string, opts DigOpts) error {
	if multiModules, err := w.DepMgr.IsMultiModules(); err != nil {
		return fmt.Errorf("failed to check if multi-modules: %w", err)
	} else if multiModules {
		return fmt.Errorf("dig does not work on the parent project. Please run it on the child project")
	}

	depTree, err := w.DepMgr.DependencyTree()
	if err != nil {
		return err
	}

	subtree, cve, found := depTree.CriticalOrHigh()
	if !found {
		fmt.Println("Congratulations! Your project has no Critical or High CVE.")
		return nil
	}

	fmt.Printf("%s found in %+v\n", cve.Cve, subtree.Root())

	return w.Kill([]string{cve.Cve}, opts.KillOpts)
}
