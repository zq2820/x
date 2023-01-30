// Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
// Use of this document is governed by a license found in the LICENSE document.

/*

Package react is a set of GopherJS bindings for Facebook's React, a Javascript
library for building user interfaces.

For more information see https://github.com/myitcv/x/blob/master/react/_doc/README.md

*/
package react

//go:generate gobin -m -run myitcv.io/react/cmd/cssGen
//go:generate gobin -m -run myitcv.io/react/cmd/coreGen

import (
	"fmt"
	"reflect"

	"github.com/gopherjs/gopherjs/chunks"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"

	// imported for the side effect of bundling react
	// build tags control whether this actually includes
	// js files or not

	_ "myitcv.io/react/internal/bundle"
	"myitcv.io/react/internal/core"
)

const (
	reactCompProps                = "props"
	reactCompLastState            = "__lastState"
	reactComponentBuilder         = "__componentBuilder"
	reactCompDisplayName          = "displayName"
	reactCompSetState             = "setState"
	reactCompForceUpdate          = "forceUpdate"
	reactCompState                = "state"
	reactCompGetInitialState      = "getInitialState"
	reactCompComponentDidMount    = "componentDidMount"
	reactGetSnapshotBeforeUpdate  = "getSnapshotBeforeUpdate"
	reactShouldComponentUpdate    = "shouldComponentUpdate"
	reactComponentDidUpdate       = "componentDidUpdate"
	reactCompComponentWillUnmount = "componentWillUnmount"
	reactCompRender               = "render"
	reactComponentDidCatch        = "componentDidCatch"
	reactGetDerivedStateFromError = "getDerivedStateFromError"
	reactGetDerivedStateFromProps = "getDerivedStateFromProps"

	nestedChildren         = "_children"
	nestedProps            = "_props"
	nestedState            = "_state"
	nestedComponentWrapper = "__ComponentWrapper"
)

var jsFragment = js.Reference("Fragment")

var jsCreateElement = js.Reference("createElement")

var jsCreateClass = js.Reference("createClass")

var jsUseState = js.Reference("useState")

var jsUseEffect = js.Reference("useEffect")

var jsUseRef = js.Reference("useRef")

var jsUseCallback = js.Reference("useCallback")

var jsUseMemo = js.Reference("useMemo")

var jsDOMRender = js.Reference("render")

var object = js.Global.Get("Object")

// ComponentDef is embedded in a type definition to indicate the type is a component
type ComponentDef[P Props, S State] struct {
	elem *js.Object
}

// var compMap map[string]*js.Object

// func init() {
// 	compMap = make(map[string]*js.Object)
// }

// S is the React representation of a string
type S = core.S

func Sprintf(format string, args ...interface{}) S {
	return S(fmt.Sprintf(format, args...))
}

func Sprintln[T any](args []T) S {
	_args := make([]interface{}, len(args))
	for i, val := range args {
		_args[i] = val
	}

	return S(fmt.Sprintln(_args...))
}

type ElementHolder = core.ElementHolder

type Element = core.Element

type Component interface {
	Render() Element
}

type componentWithDidMount interface {
	Component
	ComponentDidMount()
}

type componentWithDidUpdate[P Props, S State] interface {
	Component
	ComponentDidUpdate(prevProps P, prevState S, snapshot *js.Object)
}

type getSnapshotBeforeUpdate[P Props, S State] interface {
	Component
	GetSnapshotBeforeUpdate(prevProps P, prevState S) interface{}
}

type componentWithGetInitialState[S State] interface {
	Component
	GetInitialStateIntf() S
}

type componentWithWillUnmount interface {
	Component
	ComponentWillUnmount()
}

type componentDidCatch interface {
	Component
	ComponentDidCatch(info *js.Object, componentStack *js.Object)
}

type shouldComponentUpdate[P Props, S State] interface {
	Component
	ShouldComponentUpdate(prevProps P, prevState S) bool
}

type getDerivedStateFromError[S State] interface {
	GetDerivedStateFromError(errpr *js.Object) S
}
type getDerivedStateFromProps[P Props, S State] interface {
	GetDerivedStateFromProps(prevProps P, prevState S) S
}

type Props interface {
	IsProps()
	EqualsIntf(v Props) bool
}

type State interface {
	IsState()
	EqualsIntf(v State) bool
}

func (c *ComponentDef[P, S]) Props() P {
	if c.elem.Get(reactCompProps).Get(nestedProps) == js.Undefined {
		return reflect.ValueOf(nil).Interface().(P)
	}
	return unwrapValue(c.elem.Get(reactCompProps).Get(nestedProps)).(P)
}

func (c *ComponentDef[P, S]) Children() []Element {
	v := c.elem.Get(reactCompProps).Get(nestedChildren)

	if v == js.Undefined {
		return nil
	}

	return *(unwrapValue(v).(*[]Element))
}

func (c *ComponentDef[P, S]) SetState(i State) {
	rs := c.elem.Get(reactCompState)
	is := rs.Get(nestedState)

	cur := *(unwrapValue(is.Get(reactCompLastState)).(*State))

	if i.EqualsIntf(cur) {
		return
	}

	is.Set(reactCompLastState, wrapValue(&i))
	c.elem.Call(reactCompForceUpdate)
}

func (c *ComponentDef[P, S]) State() S {
	rs := c.elem.Get(reactCompState)
	is := rs.Get(nestedState)

	return *(unwrapValue(is.Get(reactCompLastState))).(*S)
}

func (c *ComponentDef[P, S]) ForceUpdate() {
	c.elem.Call(reactCompForceUpdate)
}

type ComponentBuilder[P Props, S State] func(elem ComponentDef[P, S]) Component

type HotComponent struct {
	ComponentDef[Props, State]
	module string
	render func(a *HotComponent) Element
}

func (a *HotComponent) Render() Element {
	return a.render(a)
}

func (a *HotComponent) ComponentDidMount() {
	dependencies := js.Global.Get("dependencies")
	if dependencies.Get(a.module) == js.Undefined {
		dependencies.Set(a.module, js.Global.Get("Array").New())
	}
	a.elem.Set("_comp", wrapValue(a))
	dependencies.Get(a.module).Call("push", a.elem.Get("_comp"))
}

func (a *HotComponent) ForceUpdate() {
	// if a.module != "" {
	// 	js.Debugger()
	// 	delete(compMap, a.module)
	// }
	a.ComponentDef.ForceUpdate()
}

func (a *HotComponent) ComponentWillUnmount() {
	dependencies := js.Global.Get("dependencies")
	dependencies.Get(a.module).Call("forEach", js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if a.elem.Get("_comp") == arguments[0] {
			dependencies.Get(a.module).Call("splice", arguments[1], 1)
		}
		return nil
	}))
}

func buildClassComponent[P Props, S State](buildCmp func(elem ComponentDef[P, S]) Component, pkg string, component interface{}, props P, children ...Element) Element {
	cmp := buildCmp(ComponentDef[P, S]{})
	var typ reflect.Type
	if reflect.TypeOf(component).Kind() == reflect.Ptr {
		typ = reflect.TypeOf(cmp).Elem()
	}
	var comp *js.Object
	// comp = compMap[pkg]
	// if comp == nil {
	comp = buildReactComponent(typ, buildCmp)
	// compMap[pkg] = comp
	// }

	propsWrap := object.New()
	if reflect.ValueOf(props).Interface() != nil {
		propsWrap.Set(nestedProps, wrapValue(props))
	}

	if children != nil {
		propsWrap.Set(nestedChildren, wrapValue(&children))
	}

	args := []interface{}{comp, propsWrap}

	for _, v := range children {
		args = append(args, v)
	}

	return &ElementHolder{
		Elem: jsCreateElement.Invoke(args...),
	}
}

func createElementHot[P Props, S State](component interface{}, props P, children ...Element) Element {
	var buildCmp ComponentBuilder[P, S] = func(elem ComponentDef[P, S]) Component {
		reflect.ValueOf(component).Elem().FieldByName("ComponentDef").Set(reflect.ValueOf(elem))
		return reflect.ValueOf(component).Interface().(Component)
	}
	return buildClassComponent(buildCmp, reflect.TypeOf(component).Elem().PkgPath(), component, props, children...)
}

func CreateElement[P Props](component interface{}, props P, children ...Element) Element {
	componentType := reflect.TypeOf(component)
	if componentType.Kind() == reflect.Ptr {
		if _, ok := reflect.TypeOf(component).Elem().FieldByName("ComponentDef"); ok {
			pkg := reflect.TypeOf(component).Elem().PkgPath()
			originalPkg := pkg
			_ = originalPkg
			var buildCmp ComponentBuilder[Props, State] = func(elem ComponentDef[Props, State]) Component {
				if chunks.IsWatch {
					hot := &HotComponent{ComponentDef: elem, module: originalPkg}
					hot.render = func(a *HotComponent) Element {
						return createElementHot[Props, State](chunks.GoChunks[originalPkg], a.Props(), a.Children()...)
					}
					return reflect.ValueOf(hot).Interface().(Component)
				}
				reflect.ValueOf(component).Elem().FieldByName("ComponentDef").Set(reflect.ValueOf(elem))
				return reflect.ValueOf(component).Interface().(Component)
			}

			if chunks.IsWatch {
				pkg += "$hot"
			}

			return buildClassComponent[Props](buildCmp, pkg, component, props, children...)
		} else {
			panic("element type error")
		}
	} else if componentType.Kind() == reflect.Func {
		if componentType.NumOut() == 1 {
			returnType := componentType.Out(0)
			// if reutrn type is Element
			if returnType.PkgPath() == "myitcv.io/react/internal/core" {
				if returnType.Name() == "Element" {
					if chunks.IsWatch {
						var buildCmp ComponentBuilder[Props, State] = func(elem ComponentDef[Props, State]) Component {
							hot := &HotComponent{ComponentDef: elem, module: componentType.PkgPath()}
							hot.render = func(a *HotComponent) Element {
								children := a.Children()
								args := make([]interface{}, 0, len(children)+2)
								args = append(args, chunks.GoChunks[componentType.PkgPath()])
								args = append(args, elem.elem.Get(reactCompProps))
								args = append(args, children)

								return &ElementHolder{
									Elem: jsCreateElement.Invoke(args...),
								}
							}
							return reflect.ValueOf(hot).Interface().(Component)
						}

						return buildClassComponent[Props](buildCmp, componentType.PkgPath()+"$hot", component, props, children...)
					}
					return CreateFunctionElement(component, props, children...)
				}
			}
		}
	}

	return nil
}

func createElement(cmp interface{}, props interface{}, children ...Element) Element {
	args := []interface{}{cmp, props}

	for _, v := range children {
		args = append(args, v)
	}

	return &ElementHolder{
		Elem: jsCreateElement.Invoke(args...),
	}
}

func buildReactComponent[P Props, S State](typ reflect.Type, builder ComponentBuilder[P, S]) *js.Object {
	compDef := object.New()
	compDef.Set("statics", object.New())
	if typ != nil {
		compDef.Set(reactCompDisplayName, fmt.Sprintf("%v(%v)", typ.Name(), typ.PkgPath()))
	} else {
		compDef.Set(reactCompDisplayName, "HotComponent(myitcv.io/react)")
	}
	compDef.Set(reactComponentBuilder, builder)

	compDef.Set(reactCompGetInitialState, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		var wv *js.Object

		res := object.New()
		is := object.New()

		if cmp, ok := cmp.(componentWithGetInitialState[S]); ok {
			x := cmp.GetInitialStateIntf()
			wv = wrapValue(&x)
		}

		res.Set(nestedState, is)
		is.Set(reactCompLastState, wv)

		return res
	}))

	compDef.Set(reactCompComponentDidMount, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		if cmp, ok := cmp.(componentWithDidMount); ok {
			cmp.ComponentDidMount()
		}

		return nil
	}))

	compDef.Set(reactComponentDidUpdate, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		if cmp, ok := cmp.(componentWithDidUpdate[P, S]); ok {
			prevProps := unwrapValue(arguments[0].Get(nestedProps)).(P)
			prevState := *unwrapValue(arguments[1].Get(nestedState).Get(reactCompLastState)).(*S)
			cmp.ComponentDidUpdate(prevProps, prevState, arguments[2])
		}

		return nil
	}))

	compDef.Set(reactShouldComponentUpdate, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		if cmp, ok := cmp.(shouldComponentUpdate[P, S]); ok {
			prevProps := unwrapValue(arguments[0].Get(nestedProps)).(P)
			prevState := *unwrapValue(arguments[1].Get(nestedState).Get(reactCompLastState)).(*S)
			return wrapValue(cmp.ShouldComponentUpdate(prevProps, prevState))
		}

		return nil
	}))

	compDef.Set(reactGetSnapshotBeforeUpdate, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		if cmp, ok := cmp.(getSnapshotBeforeUpdate[P, S]); ok {
			prevProps := unwrapValue(arguments[0].Get(nestedProps)).(P)
			prevState := *unwrapValue(arguments[1].Get(nestedState).Get(reactCompLastState)).(*S)
			return wrapValue(cmp.GetSnapshotBeforeUpdate(prevProps, prevState))
		}

		return nil
	}))

	compDef.Set(reactCompComponentWillUnmount, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		if cmp, ok := cmp.(componentWithWillUnmount); ok {
			cmp.ComponentWillUnmount()
		}

		return nil
	}))

	compDef.Set(reactComponentDidCatch, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		if cmp, ok := cmp.(componentDidCatch); ok {
			cmp.ComponentDidCatch(arguments[0], arguments[1])
		}

		return nil
	}))

	instance := builder(ComponentDef[P, S]{elem: nil})
	if cmp, ok := instance.(getDerivedStateFromError[S]); ok {
		compDef.Get("statics").Set(reactGetDerivedStateFromError, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
			var wv *js.Object
			res := object.New()
			is := object.New()
			x := cmp.GetDerivedStateFromError(arguments[0])
			wv = wrapValue(&x)
			res.Set(nestedState, is)
			is.Set(reactCompLastState, wv)

			return res
		}))
	}
	if cmp, ok := instance.(getDerivedStateFromProps[P, S]); ok {
		compDef.Get("statics").Set(reactGetDerivedStateFromProps, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
			var prevProps = unwrapValue(arguments[0].Get(nestedProps)).(P)
			var prevState = *unwrapValue(arguments[1].Get(nestedState).Get(reactCompLastState)).(*S)

			res := object.New()
			is := object.New()
			x := cmp.GetDerivedStateFromProps(prevProps, prevState)
			res.Set(nestedState, is)
			is.Set(reactCompLastState, wrapValue(&x))

			return res
		}))
	}

	compDef.Set(reactCompRender, js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		elem := this
		cmp := builder(ComponentDef[P, S]{elem: elem})

		renderRes := cmp.Render()

		return renderRes
	}))

	return jsCreateClass.Invoke(compDef)
}

func Render(el Element, container dom.Element) Element {
	v := jsDOMRender.Invoke(el, container)
	// compMap = make(map[string]*js.Object)

	return &ElementHolder{Elem: v}
}

func CreateFunctionElement[P Props](cmp interface{}, props P, children ...Element) Element {
	propsWrap := object.New()
	if reflect.ValueOf(props).Interface() != nil {
		propsWrap.Set(nestedProps, wrapValue(props))
	}

	if children != nil {
		propsWrap.Set(nestedChildren, wrapValue(&children))
	}

	args := []interface{}{makeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		unwrapChildren := reflect.ValueOf(*(unwrapValue(arguments[0].Get(nestedChildren)).(*[]Element)))
		unwrapArgs := make([]reflect.Value, 0, unwrapChildren.Len()+1)
		unwrapArgs = append(unwrapArgs, reflect.ValueOf(unwrapValue(arguments[0].Get(nestedProps))))
		for i := 0; i < unwrapChildren.Len(); i++ {
			unwrapArgs = append(unwrapArgs, unwrapChildren.Index(i))
		}

		return reflect.ValueOf(cmp).Call(unwrapArgs)[0].Interface().(Element)
	}, js.InternalObject(reflect.ValueOf(cmp)).Call("pointer").Get("name").String()), propsWrap}

	for _, v := range children {
		args = append(args, v)
	}

	return &ElementHolder{
		Elem: jsCreateElement.Invoke(args...),
	}
}

func UseState[T any](vals ...T) (T, func(T)) {
	args := make([]interface{}, 0, len(vals))
	for _, val := range vals {
		args = append(args, wrapValue(val))
	}

	v := jsUseState.Invoke(args...)

	return unwrapValue(v.Index(0)).(T), func(val T) {
		v.Index(1).Invoke(wrapValue(val))
	}
}

func UseEffect(cb func() func(), deps []interface{}) {
	jsUseEffect.Invoke(cb, deps)
}

func UseCallback[T any](cb T, deps []interface{}) T {
	return unwrapValue(jsUseCallback.Invoke(wrapValue(cb), deps)).(T)
}

func UseMemo[T any](cb func() T, deps []interface{}) T {
	f := js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		return wrapValue(cb())
	})

	return unwrapValue(jsUseMemo.Invoke(f, deps)).(T)
}

func UseRef(val ...interface{}) *js.Object {
	v := jsUseRef.Invoke(val...)

	return v
}

type FunctionComponent[P Props] interface {
	HackRender(props *js.Object) Element
	Default(props P, children ...Element) Element
}

func makeFunc(fn func(this *js.Object, arguments []*js.Object) interface{}, name string) *js.Object {
	return js.Global.Call("$makeFunc", js.InternalObject(fn), name)
}
