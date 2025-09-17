package gonfiguration

import "maps"

func SetDefault(key string, val any) {
	gonfig.setDefault(key, val)
}

func SetDefaults(defaults map[string]any) {
	gonfig.setDefaults(defaults)
}

func GetDefaults() map[string]any {
	return gonfig.getDefaults()
}

func (g *gonfiguration) setDefault(key string, val any) {
	g.Lock()
	defer g.Unlock()

	g.defaults[key] = val
}

func (g *gonfiguration) getDefault(key string) any {
	g.RLock()
	defer g.RUnlock()

	if _, ok := g.defaults[key]; !ok {
		return nil
	}

	return g.defaults[key]
}

func (g *gonfiguration) setDefaults(defaults map[string]any) {
	for key, val := range defaults {
		g.setDefault(key, val)
	}
}

func (g *gonfiguration) getDefaults() map[string]any {
	g.RLock()
	defer g.RUnlock()

	defaultsCopy := make(map[string]any, len(g.defaults))
	maps.Copy(defaultsCopy, g.defaults)

	return defaultsCopy
}
