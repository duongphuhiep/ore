package ore

import (
	"context"
)

var (
	DefaultContainer = NewContainer()

	//contextKeysRepositoryID is a special context key. The value of this key is the collection of other context keys stored in the context.
	contextKeysRepositoryID specialContextKey = "The context keys repository"
	//contextKeyResolversStack is a special context key. The value of this key is the [ResolversStack].
	contextKeyResolversStack specialContextKey = "Dependencies stack"
)

type contextKeysRepository = []contextKey

type Creator[T any] interface {
	New(ctx context.Context) (T, context.Context)
}

// Generates a unique identifier for a service resolver based on type and key(s)
func getTypeID(pointerTypeName pointerTypeName, key ...KeyStringer) typeID {
	for _, stringer := range key {
		if stringer == nil {
			panic(nilKey)
		}
	}
	return typeID{pointerTypeName, oreKey(key...)}
}

// Generates a unique identifier for a service resolver based on type and key(s)
func typeIdentifier[T any](key ...KeyStringer) typeID {
	return getTypeID(getPointerTypeName[T](), key...)
}

// Appends a service resolver to the container with type and key
func addResolver[T any](this *Container, resolver serviceResolverImpl[T], key ...KeyStringer) {
	if this.isBuilt {
		panic(alreadyBuiltCannotAdd)
	}

	typeID := typeIdentifier[T](key...)

	this.lock.Lock()
	resolver.id = contextKey{
		typeID:      typeID,
		containerID: this.containerID,
		index:       len(this.resolvers[typeID]),
	}
	this.resolvers[typeID] = append(this.resolvers[typeID], resolver)
	this.lock.Unlock()
}

func replaceResolver[T any](this *Container, resolver serviceResolverImpl[T]) {
	this.lock.Lock()
	this.resolvers[resolver.id.typeID][resolver.id.index] = resolver
	this.lock.Unlock()
}

func addAliases[TInterface, TImpl any](this *Container) {
	originalType := getPointerTypeName[TImpl]()
	aliasType := getPointerTypeName[TInterface]()
	if originalType == aliasType {
		return
	}
	this.lock.Lock()
	for _, ot := range this.aliases[aliasType] {
		if ot == originalType {
			return //already registered
		}
	}
	this.aliases[aliasType] = append(this.aliases[aliasType], originalType)
	this.lock.Unlock()
}

func Build() {
	DefaultContainer.Build()
}

func Validate() {
	DefaultContainer.Validate()
}
