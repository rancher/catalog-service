package parse

import (
	"strconv"
	"strings"
)

func TemplateURLPath(path string) (string, string, string, int, bool) {
	pathSplit := strings.Split(path, ":")
	switch len(pathSplit) {
	case 2:
		catalog := pathSplit[0]
		template := pathSplit[1]
		templateSplit := strings.Split(template, "*")
		templateBase := ""
		switch len(templateSplit) {
		case 1:
			template = templateSplit[0]
		case 2:
			templateBase = templateSplit[0]
			template = templateSplit[1]
		default:
			return "", "", "", 0, false
		}
		return catalog, template, templateBase, -1, true
	case 3:
		catalog := pathSplit[0]
		template := pathSplit[1]
		revision, err := strconv.Atoi(pathSplit[2])
		if err != nil {
			return "", "", "", 0, false
		}
		templateSplit := strings.Split(template, "*")
		templateBase := ""
		switch len(templateSplit) {
		case 1:
			template = templateSplit[0]
		case 2:
			templateBase = templateSplit[0]
			template = templateSplit[1]
		default:
			return "", "", "", 0, false
		}
		return catalog, template, templateBase, revision, true
	default:
		return "", "", "", 0, false
	}
}

func TemplatePath(path string) (string, string, bool) {
	split := strings.Split(path, "/")
	if len(split) < 2 {
		return "", "", false
	}
	return split[0], split[1], true
}

func VersionPath(path string) (string, string, int, bool) {
	split := strings.Split(path, "/")
	if len(split) < 3 {
		return "", "", 0, false
	}

	revision, err := strconv.Atoi(split[2])
	if err != nil {
		return "", "", 0, false
	}

	return split[0], split[1], revision, true
}
