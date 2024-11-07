package ore

import (
	"context"
	"sync"
)

var (
	DisableValidation = false
	lock              = &sync.RWMutex{}
	isBuilt           = false
	container         = map[typeID][]serviceResolver{}

	//map the alias type (usually an interface) to the original types (usually implementations of the interface)
	aliases = map[pointerTypeName][]pointerTypeName{}

	//contextKeysRepositoryID is a special context key. The value of this key is the collection of other context keys stored in the context.
	contextKeysRepositoryID specialContextKey = "The context keys repository"
	//contextKeyResolversChain is a special context key. The value of this key is the [ResolversChain].
	contextKeyResolversChain specialContextKey = "Dependencies chain"
)

type contextKeysRepository = []contextKey

type Creator[T any] interface {
	New(ctx context.Context) (T, context.Context)
}

// Generates a unique identifier for a service resolver based on type and key(s)
func getTypeID(pointerTypeName pointerTypeName, key []KeyStringer) typeID {
	for _, stringer := range key {
		if stringer == nil {
			panic(nilKey)
		}
	}
	return typeID{pointerTypeName, oreKey(key)}
}

// Generates a unique identifier for a service resolver based on type and key(s)
func typeIdentifier[T any](key []KeyStringer) typeID {
	return getTypeID(getPointerTypeName[T](), key)
}

// Appends a service resolver to the container with type and key
func appendToContainer[T any](resolver serviceResolverImpl[T], key []KeyStringer) {
	if isBuilt {
		panic(alreadyBuiltCannotAdd)
	}

	typeID := typeIdentifier[T](key)

	lock.Lock()
	resolver.id = contextKey{typeID, len(container[typeID])}
	container[typeID] = append(container[typeID], resolver)
	lock.Unlock()
}

func replaceServiceResolver[T any](resolver serviceResolverImpl[T]) {
	lock.Lock()
	container[resolver.id.typeID][resolver.id.index] = resolver
	lock.Unlock()
}

func appendToAliases[TInterface, TImpl any]() {
	originalType := getPointerTypeName[TImpl]()
	aliasType := getPointerTypeName[TInterface]()
	if originalType == aliasType {
		return
	}
	lock.Lock()
	for _, ot := range aliases[aliasType] {
		if ot == originalType {
			return //already registered
		}
	}
	aliases[aliasType] = append(aliases[aliasType], originalType)
	lock.Unlock()
}

func Build() {
	if isBuilt {
		panic(alreadyBuilt)
	}

	isBuilt = true
}

// Validate invokes all registered resolvers. It panics if any of them fails.
// It is recommended to call this function on application start, or in the CI/CD test pipeline
// The objectif is to panic early when the container is bad configured. For eg:
//
//   - (1) Missing depedency (forget to register certain resolvers)
//   - (2) cyclic dependency
//   - (3) lifetime misalignment (a longer lifetime service depends on a shorter one).
func Validate() {

	ctx := context.Background()
	for _, resolvers := range container {
		for _, resolver := range resolvers {
			_, ctx = resolver.resolveService(ctx)
		}
	}
}

// Reset the container, remove all the registered resolvers. For test only
func Reset() {
	clearAll()
}
