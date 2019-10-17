package pmy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/mattn/go-zglob"
)

const pmyRuleSuffix = "pmy_rules.json"

// RuleFile represents one Rule Json file
// information
type RuleFile struct {
	Path     string
	Basename string
}

func (rf RuleFile) loadRules() (Rules, error) {
	jsonFile, err := os.Open(rf.Path)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	var rules Rules
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (rf RuleFile) isApplicable(cmd string) bool {
	if rf.Basename == pmyRuleSuffix {
		return true
	}
	if rf.Basename == fmt.Sprintf(
		"%s_%s",
		cmd,
		pmyRuleSuffix,
	) {
		return true
	}
	return false
}

// GetAllRuleFiles get all pmy rules json paths
// configured by environment variable
func GetAllRuleFiles() []*RuleFile {
	ruleRoots := []string{defaultRulePath}
	ruleRoots = append(ruleRoots, strings.Split(RulePath, ":")...)
	res := []*RuleFile{}
	for _, ruleRoot := range ruleRoots {
		globPattern := fmt.Sprintf(
			`%v/**/*%v`,
			os.ExpandEnv(ruleRoot),
			pmyRuleSuffix,
		)
		matches, err := zglob.Glob(globPattern)
		if err != nil {
			panic(err)
		}
		for _, rulePath := range matches {
			res = append(
				res,
				&RuleFile{
					Path:     rulePath,
					Basename: path.Base(rulePath),
				},
			)
		}

	}
	return res
}

// Rule is a struct representing one rule
type Rule struct {
	Name           string    `json:"name"`
	Matcher        string    `json:"matcher"`
	Description    string    `json:"description"`
	RegexpLeft     string    `json:"regexpLeft"`
	RegexpRight    string    `json:"regexpRight"`
	CmdGroups      CmdGroups `json:"cmdGroups"`
	FuzzyFinderCmd string    `json:"fuzzyFinderCmd"`
	BufferLeft     string    `json:"bufferLeft"`
	BufferRight    string    `json:"bufferRight"`
	paramMap       map[string]string
}

// Rules represents slice of `Rule` struct
type Rules []*Rule

// match check if the query buffers(left and right) satisfies the concerned
// rule. if the rule regexp contains parametrized subgroups, this function expand
// the `Command` to `CommandExpanded`, where all parametrized variables were expanded.
func (rule *Rule) match(bufferLeft string, bufferRight string) (bool, error) {
	re, err := regexp.Compile(rule.RegexpLeft)
	if err != nil {
		return false, err
	}
	matches := re.FindStringSubmatch(bufferLeft)
	names := re.SubexpNames()
	if len(matches) != len(names) {
		return false, nil
	}
	paramMap := map[string]string{}
	for i, name := range names {
		if i != 0 && name != "" {
			paramMap[name] = matches[i]
		}
	}
	rule.BufferLeft = strings.Replace(
		rule.BufferLeft,
		"[]",
		bufferLeft,
		-1,
	)
	rule.BufferRight = strings.Replace(
		rule.BufferRight,
		"[]",
		bufferRight,
		-1,
	)
	rule.paramMap = paramMap
	return true, nil
}
