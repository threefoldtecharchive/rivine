package daemon

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/pflag"
	"github.com/threefoldtech/rivine/build"
)

// all modules that ship with Rivine
var (
	GatewayModule = &Module{
		Name: "Gateway",
		Description: `The gateway maintains a peer to peer connection to the network and
enables other modules to perform RPC calls on peers.
The gateway is required by all other modules.`,
	}

	ConsensusSetModule = &Module{
		Name: "Consensus Set",
		Description: `The consensus set manages everything related to consensus and keeps the
blockchain in sync with the rest of the network.`,
		Dependencies: ForceNewIdentifierSet(
			GatewayModule.Identifier(),
		),
	}

	TransactionPoolModule = &Module{
		Name:        "Transaction Pool",
		Description: `The transaction pool manages unconfirmed transactions.`,
		Dependencies: ForceNewIdentifierSet(
			ConsensusSetModule.Identifier(),
			GatewayModule.Identifier(),
		),
	}

	WalletModule = &Module{
		Name:        "Wallet",
		Description: `The wallet stores and manages coins and blockstakes.`,
		Dependencies: ForceNewIdentifierSet(
			ConsensusSetModule.Identifier(),
			TransactionPoolModule.Identifier(),
		),
	}

	BlockCreatorModule = &Module{
		Name: "Block Creator",
		Description: `The block creator participates in the proof of block stake protocol
for creating new blocks. BlockStakes are required to participate.`,
		Dependencies: ForceNewIdentifierSet(
			ConsensusSetModule.Identifier(),
			TransactionPoolModule.Identifier(),
			WalletModule.Identifier(),
		),
	}

	ExplorerModule = &Module{
		Name: "Explorer",
		Description: `The explorer provides statistics about the blockchain and can be
queried for information about specific transactions or other objects on
the blockchain.`,
		Dependencies: ForceNewIdentifierSet(
			ConsensusSetModule.Identifier(),
		),
	}
)

// DefaultModuleSetFlag returns a new ModuleSetFlag,
// using the default available modules, flag names and a Rivine-defined default module set.
func DefaultModuleSetFlag() ModuleSetFlag {
	msFlag, err := NewModuleSetFlag("modules", "M",
		ForceNewIdentifierSet(
			ConsensusSetModule.Identifier(),
			GatewayModule.Identifier(),
			TransactionPoolModule.Identifier(),
			WalletModule.Identifier(),
			BlockCreatorModule.Identifier(),
		),
		DefaultModuleSet())
	if err != nil {
		build.Critical(err)
	}
	return msFlag
}

// DefaultModuleSet returns the default module set,
// containing all the modules that ship with Rivine.
func DefaultModuleSet() ModuleSet {
	set, err := NewModuleSet(
		GatewayModule,
		ConsensusSetModule,
		TransactionPoolModule,
		WalletModule,
		BlockCreatorModule,
		ExplorerModule,
	)
	if err != nil {
		build.Critical(err)
	}
	return set
}

type (
	// Module is the type of a module that can be used in combination
	// with other modules in order to form a client, whether or not as a daemon.
	//
	// Module is NOT thread-safe.
	Module struct {
		// Name of the Module (required)
		Name string
		// Description of the Module (required)
		Description string
		// Dependencies of a module (optional),
		// if given, all listed dependencies have to be unqiue
		// and refernce an existing/available module.
		Dependencies ModuleIdentifierSet
	}

	// ModuleSet is the type which represents a unique set of Modules.
	//
	// ModuleSet is NOT thread-safe.
	ModuleSet struct {
		modules []Module
	}
	// ModuleIdentifier is the type used to represent a Module's identifier
	ModuleIdentifier rune
	// ModuleIdentifierSet is a string which can contain only unique
	// and valid Module Identifiers.
	//
	// ModuleIdentifierSet is NOT thread-safe.
	ModuleIdentifierSet struct {
		identifiers []ModuleIdentifier
	}

	// ModuleSetFlag can be used as a CLI flag within a Cobra-made CLI client,
	// as a way to define a set of modules by their identifier,
	// given a set of available modules, a default identifier set and the
	// flag name which is used.
	//
	// ModuleSetFlag is NOT thread-safe.
	ModuleSetFlag struct {
		availableModules                ModuleSet
		identifiers, defaultIdentifiers ModuleIdentifierSet
		longFlag, shortFlag             string
	}
)

// NewModuleSetFlag creates a new module set flag
func NewModuleSetFlag(longFlag, shortFlag string, defaultIdentifiers ModuleIdentifierSet, availableModules ModuleSet) (flag ModuleSetFlag, err error) {
	if len(longFlag) == 0 {
		err = errors.New("no long flag given")
		return
	}
	flag.longFlag = longFlag
	if len(shortFlag) > 1 {
		err = errors.New("short flag can only be one byte long")
		return
	}
	flag.shortFlag = shortFlag
	flag.availableModules = availableModules
	flag.defaultIdentifiers = defaultIdentifiers
	// validate that all dependencies of the default identifiers can be satisfied
	err = flag.availableModules.ValidateIdentifierSet(flag.defaultIdentifiers)
	return // could return an error
}

// RegisterFlag registers this ModuleSetFlag to a given flag set,
// optionally informing the user about a help command devoted as a manual for this flag.
func (msf *ModuleSetFlag) RegisterFlag(set *pflag.FlagSet, helpCommand string) {
	description := "define the enabled modules (available modules: " +
		msf.availableModules.String() + ")"
	if helpCommand != "" {
		description += " use '" + helpCommand + "' for more information"
	}
	if msf.shortFlag != "" {
		set.VarP(msf, msf.longFlag, msf.shortFlag, description)
		return
	}
	set.Var(msf, msf.longFlag, description)
}

// String returns all module identifiers,
// either the default ones, or the one set by the user
func (msf *ModuleSetFlag) String() string {
	set := msf.ModuleIdentifiers()
	return set.String()
}

// Set resets the registered set of identifiers and tries
// to append each byte of the given string as a unique module identifier.
func (msf *ModuleSetFlag) Set(str string) (err error) {
	msf.identifiers.identifiers = nil // reset identifiers
	for _, id := range str {
		// append all module identifiers in the given order
		err = msf.AppendModuleIdentifier(ModuleIdentifier(id))
		if err != nil {
			return
		}
	}
	return msf.availableModules.ValidateIdentifierSet(msf.identifiers)
}

// AppendModuleIdentifier appends a given identifier to the registered set of identifiers
func (msf *ModuleSetFlag) AppendModuleIdentifier(id ModuleIdentifier) error {
	for _, mod := range msf.availableModules.modules {
		if mod.Identifier() != id {
			continue
		}
		return msf.identifiers.Append(id)
	}
	return fmt.Errorf(
		"%s is not linked to an available module, it has to be one of: %s",
		string(id), msf.availableModules.String())
}

// ModuleIdentifiers returns a copy of either the default identifier set,
// or if registered a copy of the set of registered identifiers.
func (msf *ModuleSetFlag) ModuleIdentifiers() ModuleIdentifierSet {
	set := msf.moduleIdentifiers()
	return ModuleIdentifierSet{identifiers: set.Identifiers()}
}

// moduleIdentifiers returns either the default identifier set,
// or if registered the set of registered identifiers.
func (msf *ModuleSetFlag) moduleIdentifiers() ModuleIdentifierSet {
	if len(msf.identifiers.identifiers) != 0 {
		return msf.identifiers
	}
	return msf.defaultIdentifiers
}

// Type returns the flag type in string format of this flag.
func (msf *ModuleSetFlag) Type() string {
	return "Modules Set Flag"
}

// WriteDescription writes a stringified description of
// this flag, including a description for each available module.
func (msf ModuleSetFlag) WriteDescription(w io.Writer) error {
	flagName := "--" + msf.longFlag
	if len(msf.shortFlag) > 0 {
		flagName += "/-" + msf.shortFlag
	}
	_, err := w.Write([]byte(fmt.Sprintf(`Use the %[1]s flag to only run specific modules. Modules are
independent components. This flag should only be used by developers or
people who want to reduce overhead from unused modules. Modules are specified by
their first letter.

If the %[1]s flag is not specified the default modules are used:
`, flagName)))
	if err != nil {
		return fmt.Errorf("failed to write initial paragraphs of the module set flag description: %v", err)
	}

	for _, id := range msf.defaultIdentifiers.identifiers {
		_, err = w.Write([]byte("  * " + msf.availableModules.moduleForIdentifier(id).Name + " (" + string(id) + ")\r\n"))
		if err != nil {
			return fmt.Errorf("failed to write default identifier %s: %v", string(id), err)
		}
	}
	_, err = w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write extra newline prior to description of available modules: %v", err)
	}
	err = msf.availableModules.WriteDescription(w)
	if err != nil {
		return fmt.Errorf("failed to write description of available modules: %v", err)
	}
	return nil
}

// ForceNewIdentifierSet creates an identifier set with the given unique identifiers,
// panicking if a given identifier is invalid or not unique.
func ForceNewIdentifierSet(identifiers ...ModuleIdentifier) ModuleIdentifierSet {
	set, err := NewIdentifierSet(identifiers...)
	if err != nil {
		build.Critical("failed to create new identifier set: " + err.Error())
	}
	return set
}

// NewIdentifierSet creates an identifier set with the given unique identifiers,
// returning an error if a given identifier is invalid or not unique.
func NewIdentifierSet(identifiers ...ModuleIdentifier) (set ModuleIdentifierSet, err error) {
	if len(identifiers) == 0 {
		return ModuleIdentifierSet{}, errors.New("no identifiers given to create a set from")
	}
	for _, id := range identifiers {
		err = set.Append(id)
		if err != nil {
			return
		}
	}
	return
}

// Length returns the length of the module identifier set.
func (set ModuleIdentifierSet) Len() int {
	return len(set.identifiers)
}

// Less implemenets sort.Interface.Less
func (set ModuleIdentifierSet) Less(i, j int) bool {
	return set.identifiers[i] < set.identifiers[j]
}

// Swap implemenets sort.Interface.Swap
func (set ModuleIdentifierSet) Swap(i, j int) {
	set.identifiers[i], set.identifiers[j] = set.identifiers[j], set.identifiers[i]
}

// Contains returns True if the set contains the given id.
func (set ModuleIdentifierSet) Contains(id ModuleIdentifier) bool {
	for _, sid := range set.identifiers {
		if sid == id {
			return true
		}
	}
	return false
}

// Identifiers returns a copy of the internal identifier slice.
func (set ModuleIdentifierSet) Identifiers() (ids []ModuleIdentifier) {
	ids = make([]ModuleIdentifier, len(set.identifiers))
	copy(ids[:], set.identifiers[:])
	return
}

// String returns the identifier set as a string,
// one byte per identifier in this set.
//
// An empty string can be returned in case the identifier set is empty.
func (set ModuleIdentifierSet) String() (str string) {
	for _, id := range set.identifiers {
		str += string(id)
	}
	return
}

// Append appends a unique module identifier to this identifier set,
// returning an error in case the given identifier (byte) is not valid or not unique.
func (set *ModuleIdentifierSet) Append(id ModuleIdentifier) error {
	id = ModuleIdentifier(unicode.ToLower(rune(id)))
	if !IsValidModuleIdentifier(id) {
		return fmt.Errorf("invalid module identifier %q", string(id))
	}
	for _, sID := range set.identifiers {
		if ModuleIdentifier(sID) == id {
			return fmt.Errorf("identifier set already contains identifier %s", string(id))
		}
	}
	set.identifiers = append(set.identifiers, id)
	return nil
}

// AppendIfUnique appends a unique module identifier to this identifier set,
// returning an error in case the given identifier (byte) is not valid,
// and returning false if the given identifier (byte) If it is not unique.
func (set *ModuleIdentifierSet) AppendIfUnique(id ModuleIdentifier) (bool, error) {
	id = ModuleIdentifier(unicode.ToLower(rune(id)))
	if !IsValidModuleIdentifier(id) {
		return false, fmt.Errorf("invalid module identifier %q", string(id))
	}
	for _, sID := range set.identifiers {
		if ModuleIdentifier(sID) == id {
			return false, nil
		}
	}
	set.identifiers = append(set.identifiers, id)
	return true, nil
}

// Difference returns the symmetric difference set of symmetric difference of this set and the other set.
// Symmetric difference meaning that it will return a new set containing the elements which are in this set,
// but not in the other set, as well as elements which are in the other set but not in this set.
func (set ModuleIdentifierSet) Difference(other ModuleIdentifierSet) (c ModuleIdentifierSet) {
	// copy internal slices and sort them
	a := ModuleIdentifierSet{identifiers: set.Identifiers()}
	b := ModuleIdentifierSet{identifiers: other.Identifiers()}
	sort.Sort(a)
	sort.Sort(b)

	lengthA, lengthB := a.Len(), b.Len()
	var indexA, indexB int
	for indexA < lengthA && indexB < lengthB {
		if a.identifiers[indexA] == b.identifiers[indexB] {
			indexA++
			indexB++
			continue
		}
		if a.identifiers[indexA] < b.identifiers[indexB] {
			// append from the first set
			c.Append(a.identifiers[indexA])
			indexA++
			continue
		}
		// append from the second set
		c.Append(b.identifiers[indexB])
		indexB++
	}
	// append all remaining ones
	for indexA < lengthA {
		c.Append(a.identifiers[indexA])
		indexA++
	}
	for indexB < lengthB {
		c.Append(b.identifiers[indexB])
		indexB++
	}
	// sort our complement and return
	sort.Sort(c)
	return
}

// IsValidModuleIdentifier checks if a given ModuleIdentifier can be
// is valid, meaning it represents a lowercase alphabetical ASCII character.
func IsValidModuleIdentifier(id ModuleIdentifier) bool {
	return id >= 'a' && id <= 'z'
}

// NewModuleSet ensures that all given modules are valid and unique,
// returning them as a module set with no error if they are all valid indeed.
func NewModuleSet(modules ...*Module) (set ModuleSet, err error) {
	if len(modules) == 0 {
		return ModuleSet{}, errors.New("no modules given to create a set from")
	}
	var idSet ModuleIdentifierSet
	for _, mod := range modules {
		if mod.Name == "" {
			return ModuleSet{}, errors.New("name is a required property of a module")
		}
		if mod.Description == "" {
			return ModuleSet{}, fmt.Errorf(
				"description is a required property of a module: %s has none",
				mod.Name)
		}
		b := mod.Identifier()
		if !IsValidModuleIdentifier(b) {
			return ModuleSet{}, fmt.Errorf(
				"module %s has an invalid identifier '%s'",
				mod.Name, string(b))
		}
		err = idSet.Append(b)
		if err != nil {
			return ModuleSet{}, err
		}
		// append a clone of the valid and unique module
		set.modules = append(set.modules, Module{
			Name:         mod.Name,
			Description:  mod.Description,
			Dependencies: ModuleIdentifierSet{identifiers: mod.Dependencies.Identifiers()},
		})
	}
	return set, nil
}

// Append the a copy of the given module to the module set.
func (ms *ModuleSet) Append(mod *Module) error {
	if mod == nil {
		build.Critical("nil module cannot be added")
	}
	if mod.Name == "" {
		return errors.New("name is a required property of a module")
	}
	if mod.Description == "" {
		return fmt.Errorf(
			"description is a required property of a module: %s has none",
			mod.Name)
	}
	b := mod.Identifier()
	if !IsValidModuleIdentifier(b) {
		return fmt.Errorf(
			"module %s has an invalid identifier '%s'",
			mod.Name, string(b))
	}
	id := mod.Identifier()
	for _, mod := range ms.modules {
		if mod.Identifier() == id {
			return fmt.Errorf("a module with the ID %s already exists in the module set", string(id))
		}
	}
	// add the unique module
	ms.modules = append(ms.modules, Module{
		Name:         mod.Name,
		Description:  mod.Description,
		Dependencies: ModuleIdentifierSet{identifiers: mod.Dependencies.Identifiers()},
	})
	return nil
}

// Set a copy of the given module in the module set,
// overwriting an existing module if it has the same identifier as the given module.
func (ms *ModuleSet) Set(mod *Module) {
	if mod == nil {
		build.Critical("nil module cannot be set")
	}
	id := mod.Identifier()
	for idx, origMod := range ms.modules {
		if origMod.Identifier() == id {
			// overwrite an existing module
			ms.modules[idx] = Module{
				Name:         mod.Name,
				Description:  mod.Description,
				Dependencies: ModuleIdentifierSet{identifiers: mod.Dependencies.Identifiers()},
			}
			return
		}
	}
	// add the module as a new unique module
	ms.modules = append(ms.modules, Module{
		Name:         mod.Name,
		Description:  mod.Description,
		Dependencies: ModuleIdentifierSet{identifiers: mod.Dependencies.Identifiers()},
	})
	return
}

// String returns the module set as a string,
// one byte per module in this set, and where each byte the identifier of that module is.
//
// An empty string can be returned in case the module set is empty.
func (ms ModuleSet) String() (str string) {
	for _, mod := range ms.modules {
		str += string(mod.Identifier())
	}
	return
}

// WriteDescription writes a full stringified description of
// all the modules in the module set.
func (ms ModuleSet) WriteDescription(w io.Writer) error {
	_, err := w.Write([]byte("The available modules are:\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write initial sentence of the description of available modules: %v", err)
	}
	// summarize all modules
	for _, module := range ms.modules {
		_, err = w.Write([]byte("  * " + module.Name + " (" + string(module.Identifier()) + ")\r\n"))
		if err != nil {
			return fmt.Errorf("failed to write summarized module %s as part of the description of available modules: %v", module.Name, err)
		}
	}

	_, err = w.Write([]byte(`
Each module can also have a list of dependencies, meaning when you want to use a module,
you also have to specify (and thus use) Its dependencies.`))
	if err != nil {
		return fmt.Errorf("failed to write pre-paragraph of the detailed description of available modules: %v", err)
	}
	// go over all modules and print the description and dependencies
	for _, module := range ms.modules {
		_, err = w.Write([]byte("\r\n"))
		if err != nil {
			return fmt.Errorf("failed to write extra newline: %v", err)
		}

		// add the module full description
		err = module.WriteDescription(w)
		if err != nil {
			return err
		}
		if len(module.Dependencies.identifiers) == 0 {
			continue
		}
		_, err = w.Write([]byte("\r\n"))
		if err != nil {
			return fmt.Errorf("failed to write extra newline: %v", err)
		}
		_, err = w.Write([]byte("> Dependencies: "))
		if err != nil {
			return fmt.Errorf("failed to write dependency prefix for module %s: %v", module.Name, err)
		}

		// collect all modifiers first]
		var depSet ModuleIdentifierSet
		for _, dep := range module.Dependencies.identifiers {
			err = ms.createDependencySetFor(dep, &depSet)
			if err != nil {
				return fmt.Errorf("failed to resolve dependencies for %s's dependency %s: %v",
					module.Name, string(dep), err)
			}
		}
		// remove the module itself from the depset
		modID := module.Identifier()
		for idx, dep := range depSet.identifiers {
			if dep == modID {
				depSet.identifiers = append(depSet.identifiers[:idx-1], depSet.identifiers[idx+1:]...)
				break
			}
		}
		// print all found dependencies
		for idx, dep := range depSet.identifiers {
			mod := ms.moduleForIdentifier(ModuleIdentifier(dep))
			if mod == nil {
				continue
			}
			_, err = w.Write([]byte(mod.Name + " (" + string(mod.Identifier()) + ")"))
			if err != nil {
				return fmt.Errorf("failed to write dependency %s for module %s: %v", mod.Name, mod.Name, err)
			}
			if idx < len(depSet.identifiers)-1 {
				_, err = w.Write([]byte(", "))
				if err != nil {
					return fmt.Errorf("failed to write dependency seperator after dependency %s for module %s: %v", mod.Name, mod.Name, err)
				}
			}
		}
	}
	_, err = w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write extra newline: %v", err)
	}
	return nil
}

// CreateDependencySetFor creates a set of identifiers
// that the given identifier indirectly or directly references
// within the context of this Module set.
func (ms ModuleSet) CreateDependencySetFor(identifier ModuleIdentifier) (set ModuleIdentifierSet, err error) {
	err = ms.createDependencySetFor(identifier, &set)
	return
}

// createDependencySetFor is a recursive internal util method,
// to walk the dependency graph, visiting each path just once
func (ms ModuleSet) createDependencySetFor(identifier ModuleIdentifier, set *ModuleIdentifierSet) error {
	// every module depends upon itself
	unique, err := set.AppendIfUnique(identifier)
	if err != nil {
		// the given identifier isn't valid
		return fmt.Errorf("given identifier %s is not valid: %v", string(identifier), err)
	}
	if !unique {
		return nil // already seen this dependency, we can skip it
	}
	mod := ms.moduleForIdentifier(identifier)
	if mod == nil {
		return fmt.Errorf("identifier %s does not reference a module in this set", string(identifier))
	}
	for _, dep := range mod.Dependencies.identifiers {
		err = ms.createDependencySetFor(ModuleIdentifier(dep), set)
		if err != nil {
			return fmt.Errorf(
				"failed to create dependency set for dependency %s of module ID %s: %v",
				string(dep), string(identifier), err)
		}
	}
	return nil // all good
}

// util method to find the pointer to a module within this set,
// for a given identifier
func (ms ModuleSet) moduleForIdentifier(identifier ModuleIdentifier) (mod *Module) {
	for idx := range ms.modules {
		mod = &ms.modules[idx]
		if mod.Identifier() == identifier {
			return mod
		}
	}
	return nil
}

// ValidateIdentifierSet validates that all given identifiers exist
// within this module set, and that all dependencies of a referenced module
// are referenced within the given set of module identifiers.
func (ms ModuleSet) ValidateIdentifierSet(set ModuleIdentifierSet) error {
	// collect the entire required dependency tree
	var dependencySet ModuleIdentifierSet
	for _, id := range set.identifiers {
		err := ms.createDependencySetFor(ModuleIdentifier(id), &dependencySet)
		if err != nil {
			return err
		}
	}
	// ensure difference is 0, as that means all dependencies are referenced in the given set
	unresolvedDependencySet := set.Difference(dependencySet)
	if unresolvedDependencySet.Len() > 0 {
		var names []string
		for _, id := range unresolvedDependencySet.identifiers {
			names = append(names, ms.moduleForIdentifier(id).Name+"("+string(id)+")")
		}
		return errors.New("unresolved module dependencies: {" +
			strings.Join(names, ",") + "}")
	}
	return nil // all good, we could satisfy all dependencies
}

// WriteDescription writes a discription of the Module in a full detailed way.
func (m Module) WriteDescription(w io.Writer) error {
	str := fmt.Sprintf(`%s (%s):`, m.Name, string(m.Identifier()))
	_, err := w.Write([]byte(fmt.Sprintf(`
%s
%s
%s`, str, strings.Repeat("-", len(str)), m.Description)))
	if err != nil {
		return fmt.Errorf("failed to write description of module %s: %v", m.Name, err)
	}
	return nil
}

// Identifier returns the lower-cased first letter of the module's name as its identifier.
func (m Module) Identifier() ModuleIdentifier {
	return ModuleIdentifier(unicode.ToLower(rune(m.Name[0])))
}
