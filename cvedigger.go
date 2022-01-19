package woodpecker

import "fmt"

type DigOpts struct {
	KillOpts
}

func (w Woodpecker) Dig(args []string, opts DigOpts) error {
	depTree, err := w.DepMgr.DependencyTree()
	if err != nil {
		return err
	}

	subtree, target, found := depTree.CriticalOrHigh()
	if !found {
		fmt.Println("Congratulations! Your project has no Critical or High CVE.")
		return nil
	}

	fmt.Printf("%s found in %+v\n", target.ID, subtree.Root())

	if subtree.Root().Depth == 0 {
		subw := w
		dm, err := subw.DepMgr.SubModule(subtree.Root().ID)
		if err != nil {
			return fmt.Errorf("failed to get sub-module %s: %s", subtree.Root().ID, err)
		}

		subw.DepMgr = dm

		return subw.Kill([]string{target.ID}, opts.KillOpts)
	} else {
		return w.Kill([]string{target.ID}, opts.KillOpts)
	}
}
