package main

func validateDetailBlock(manifest Manifest, path string, block DetailBlock) []problem {
	switch block.Kind {
	case "paragraph":
		return requireText(path, block.Text)
	case "bullets", "ordered":
		return requireItems(path, block.Items)
	case "table":
		return validateTable(path, block.Table)
	case "env_table":
		return validateEnvNames(manifest, path, block.EnvVars)
	default:
		return []problem{{Message: "invalid detail doc block in " + path}}
	}
}

func requireText(path, text string) []problem {
	if text == "" {
		return []problem{{Message: "empty detail paragraph in " + path}}
	}
	return nil
}

func requireItems(path string, items []string) []problem {
	if len(items) == 0 {
		return []problem{{Message: "empty detail list in " + path}}
	}
	return nil
}
