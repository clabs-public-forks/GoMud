package spells

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"

	"github.com/GoMudEngine/GoMud/internal/util"
)

func CreateNewSpellFile(newSpellInfo SpellData) (string, error) {

	if sp := GetSpell(newSpellInfo.SpellId); sp != nil {
		return ``, errors.New(`Spell already exists.`)
	}

	if err := newSpellInfo.Validate(); err != nil {
		return ``, err
	}

	saveModes := []fileloader.SaveOption{}

	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*SpellData](string(configs.GetFilePathsConfig().DataFiles)+`/spells`, &newSpellInfo, saveModes...); err != nil {
		return ``, err
	}

	// Save to in-memory cache
	allSpells[newSpellInfo.Id()] = &newSpellInfo

	newScriptPath := newSpellInfo.GetScriptPath()
	if err := os.MkdirAll(filepath.Dir(newScriptPath), 0o755); err != nil {
		return ``, err
	}

	fileloader.CopyFileContents(util.FilePath(`_datafiles/sample-scripts/spells/`+string(newSpellInfo.Type)+`.js`),
		newScriptPath)

	return newSpellInfo.SpellId, nil
}
