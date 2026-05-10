package main

import (
	"errors"
	"os"
	"strings"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/database"
	"github.com/ayussh-2/internal/logger"
	"github.com/ayussh-2/internal/models"
	"github.com/ayussh-2/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type seedProblem struct {
	title       string
	difficulty  string
	topics      []string
	hints       []string
	details     string
	examples    []services.Example
	constraints []string
	samples     []services.SampleTestCases
	hidden      []services.SampleTestCases
}

func tc(input, expected string) services.SampleTestCases {
	return services.SampleTestCases{Input: input, Expected: expected}
}

func ex(input, output, explanation string) services.Example {
	return services.Example{Input: input, Output: output, Explanation: explanation}
}

func main() {
	cfg := config.LoadConfig()

	log, err := logger.Init(cfg.Env)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatal("cannot connect to db", zap.Error(err))
	}

	svc := services.NewProblemService(log, db, cfg)
	problems := seedData()

	topicIDs, err := ensureTopics(db, log, problems)
	if err != nil {
		log.Fatal("could not ensure topics", zap.Error(err))
	}
	log.Info("ensured topics", zap.Int("count", len(topicIDs)))

	inserted, skipped := 0, 0
	for _, p := range problems {
		slug := slugify(p.title)

		exists, err := problemExistsBySlug(db, slug)
		if err != nil {
			log.Error("slug lookup failed", zap.String("slug", slug), zap.Error(err))
			os.Exit(1)
		}
		if exists {
			log.Info("skipping existing problem", zap.String("slug", slug))
			skipped++
			continue
		}

		ids := make([]uint, 0, len(p.topics))
		for _, name := range p.topics {
			id, ok := topicIDs[name]
			if !ok {
				log.Fatal("missing topic id", zap.String("topic", name))
			}
			ids = append(ids, id)
		}

		created, err := svc.AddProblem(services.Problem{
			Title:           p.title,
			Difficulty:      p.difficulty,
			Topics:          ids,
			Hint:            p.hints,
			Details:         p.details,
			Examples:        p.examples,
			Constraints:     p.constraints,
			SampleTestCases: p.samples,
		})
		if err != nil {
			log.Fatal("create problem failed", zap.String("title", p.title), zap.Error(err))
		}

		if len(p.hidden) > 0 {
			if _, err := svc.ReplaceHiddenTestCases(created.Slug, p.hidden); err != nil {
				log.Fatal("hidden test cases failed", zap.String("slug", created.Slug), zap.Error(err))
			}
		}

		log.Info("seeded problem",
			zap.String("title", p.title),
			zap.String("slug", created.Slug),
			zap.Int("samples", len(p.samples)),
			zap.Int("hidden", len(p.hidden)),
		)
		inserted++
	}

	log.Info("seed complete", zap.Int("inserted", inserted), zap.Int("skipped", skipped))
}

func ensureTopics(db *gorm.DB, log *zap.Logger, problems []seedProblem) (map[string]uint, error) {
	seen := map[string]struct{}{}
	names := make([]string, 0)
	for _, p := range problems {
		for _, name := range p.topics {
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	out := map[string]uint{}
	for _, name := range names {
		topic, err := upsertTopic(db, name)
		if err != nil {
			return nil, err
		}
		out[name] = topic.ID
		log.Debug("topic ready", zap.String("name", name), zap.Uint("id", topic.ID))
	}
	return out, nil
}

func upsertTopic(db *gorm.DB, name string) (*models.Topics, error) {
	var topic models.Topics
	err := db.Where("name = ?", name).First(&topic).Error
	if err == nil {
		return &topic, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	topic = models.Topics{Name: name}
	if err := db.Create(&topic).Error; err != nil {
		return nil, err
	}
	return &topic, nil
}

func problemExistsBySlug(db *gorm.DB, slug string) (bool, error) {
	var count int64
	if err := db.Model(&models.Problems{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if b.Len() > 0 && !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "problem"
	}
	return out
}

func seedData() []seedProblem {
	return []seedProblem{
		{
			title:      "Two Sum",
			difficulty: "easy",
			topics:     []string{"Arrays", "Hashing"},
			hints: []string{
				"A naive solution is $O(n^2)$. Can you do it in one pass?",
				"Use a map from value to index to find the complement in $O(1)$.",
			},
			details: `## Statement

Given an array of integers $nums$ and an integer $target$, return the **indices** of the two numbers that add up to $target$.

You may assume each input has exactly one solution, and you may not use the same element twice.`,
			examples: []services.Example{
				ex("nums = [2,7,11,15], target = 9", "[0,1]", "nums[0] + nums[1] == 9"),
				ex("nums = [3,2,4], target = 6", "[1,2]", "nums[1] + nums[2] == 6"),
			},
			constraints: []string{
				"$2 \\le n \\le 10^4$",
				"$-10^9 \\le nums[i] \\le 10^9$",
				"Only one valid answer exists.",
			},
			samples: []services.SampleTestCases{
				tc("nums = [2,7,11,15]\ntarget = 9", "[0,1]"),
				tc("nums = [3,2,4]\ntarget = 6", "[1,2]"),
			},
			hidden: []services.SampleTestCases{
				tc("nums = [3,3]\ntarget = 6", "[0,1]"),
				tc("nums = [-1,-2,-3,-4,-5]\ntarget = -8", "[2,4]"),
				tc("nums = [0,4,3,0]\ntarget = 0", "[0,3]"),
			},
		},
		{
			title:      "Valid Parentheses",
			difficulty: "easy",
			topics:     []string{"Stack", "Strings"},
			hints: []string{
				"Walk the string and use a stack to track open brackets.",
				"On a closing bracket, the top of the stack must be the matching opener.",
			},
			details: `## Statement

Given a string $s$ containing only the characters $($, $)$, $\{$, $\}$, $[$, $]$, determine if it is **valid**.

A string is valid if open brackets are closed by the same type of bracket and in the correct order.`,
			examples: []services.Example{
				ex(`s = "()"`, "true", "Single matched pair."),
				ex(`s = "(]"`, "false", "Mismatched bracket types."),
			},
			constraints: []string{
				"$1 \\le |s| \\le 10^4$",
				"$s$ consists only of bracket characters.",
			},
			samples: []services.SampleTestCases{
				tc(`s = "()[]{}"`, "true"),
				tc(`s = "([)]"`, "false"),
			},
			hidden: []services.SampleTestCases{
				tc(`s = "{[]}"`, "true"),
				tc(`s = "("`, "false"),
				tc(`s = ""`, "true"),
			},
		},
		{
			title:      "Reverse Linked List",
			difficulty: "easy",
			topics:     []string{"Linked List"},
			hints: []string{
				"Iterate the list while flipping each node's next pointer.",
				"Keep three pointers: prev, curr, next.",
			},
			details: `## Statement

Given the $head$ of a singly linked list, reverse the list and return the new head.`,
			examples: []services.Example{
				ex("head = [1,2,3,4,5]", "[5,4,3,2,1]", "Reverse the order of nodes."),
			},
			constraints: []string{
				"$0 \\le n \\le 5000$",
				"$-5000 \\le \\text{Node.val} \\le 5000$",
			},
			samples: []services.SampleTestCases{
				tc("head = [1,2,3,4,5]", "[5,4,3,2,1]"),
				tc("head = [1,2]", "[2,1]"),
			},
			hidden: []services.SampleTestCases{
				tc("head = []", "[]"),
				tc("head = [7]", "[7]"),
				tc("head = [1,1,2,2]", "[2,2,1,1]"),
			},
		},
		{
			title:      "Maximum Subarray",
			difficulty: "medium",
			topics:     []string{"Arrays", "Dynamic Programming"},
			hints: []string{
				"Track the best sum ending at the current index.",
				"Reset the running sum when it drops below zero.",
			},
			details: `## Statement

Given an integer array $nums$, find the **contiguous** subarray with the largest sum and return its sum.

This is also known as **Kadane's algorithm**.`,
			examples: []services.Example{
				ex("nums = [-2,1,-3,4,-1,2,1,-5,4]", "6", "Subarray [4,-1,2,1] has sum 6."),
			},
			constraints: []string{
				"$1 \\le n \\le 10^5$",
				"$-10^4 \\le nums[i] \\le 10^4$",
			},
			samples: []services.SampleTestCases{
				tc("nums = [-2,1,-3,4,-1,2,1,-5,4]", "6"),
				tc("nums = [1]", "1"),
			},
			hidden: []services.SampleTestCases{
				tc("nums = [5,4,-1,7,8]", "23"),
				tc("nums = [-1,-2,-3,-4]", "-1"),
				tc("nums = [0,0,0]", "0"),
			},
		},
		{
			title:      "Climbing Stairs",
			difficulty: "easy",
			topics:     []string{"Dynamic Programming"},
			hints: []string{
				"How many ways to reach step $n$ if you can step 1 or 2?",
				"It's the Fibonacci sequence.",
			},
			details: `## Statement

You are climbing a staircase with $n$ steps. Each move you can climb $1$ or $2$ steps.

Return the number of distinct ways to reach the top.`,
			examples: []services.Example{
				ex("n = 2", "2", "1+1 or 2."),
				ex("n = 3", "3", "1+1+1, 1+2, or 2+1."),
			},
			constraints: []string{
				"$1 \\le n \\le 45$",
			},
			samples: []services.SampleTestCases{
				tc("n = 2", "2"),
				tc("n = 3", "3"),
			},
			hidden: []services.SampleTestCases{
				tc("n = 1", "1"),
				tc("n = 5", "8"),
				tc("n = 10", "89"),
			},
		},
		{
			title:      "Binary Search",
			difficulty: "easy",
			topics:     []string{"Binary Search", "Arrays"},
			hints: []string{
				"Maintain $lo$ and $hi$ pointers and halve the range each step.",
				"Use $lo + (hi - lo) / 2$ to avoid integer overflow.",
			},
			details: `## Statement

Given a sorted array $nums$ and a value $target$, return the index of $target$ if present, otherwise return $-1$.

Your algorithm must run in $O(\log n)$ time.`,
			examples: []services.Example{
				ex("nums = [-1,0,3,5,9,12], target = 9", "4", ""),
				ex("nums = [-1,0,3,5,9,12], target = 2", "-1", ""),
			},
			constraints: []string{
				"$1 \\le n \\le 10^4$",
				"$nums$ is sorted in ascending order.",
				"All values are unique.",
			},
			samples: []services.SampleTestCases{
				tc("nums = [-1,0,3,5,9,12]\ntarget = 9", "4"),
				tc("nums = [-1,0,3,5,9,12]\ntarget = 2", "-1"),
			},
			hidden: []services.SampleTestCases{
				tc("nums = [5]\ntarget = 5", "0"),
				tc("nums = [5]\ntarget = -3", "-1"),
				tc("nums = [1,2,3,4,5,6,7,8,9,10]\ntarget = 1", "0"),
			},
		},
		{
			title:      "Best Time to Buy and Sell Stock",
			difficulty: "easy",
			topics:     []string{"Arrays", "Dynamic Programming"},
			hints: []string{
				"Track the minimum price seen so far.",
				"At each day, the best profit is $price - minSoFar$.",
			},
			details: `## Statement

You are given an array $prices$ where $prices[i]$ is the price of a stock on day $i$.

Choose a single day to buy and a later day to sell so that profit is maximized. Return the maximum profit, or $0$ if no profit is achievable.`,
			examples: []services.Example{
				ex("prices = [7,1,5,3,6,4]", "5", "Buy at 1 (day 2), sell at 6 (day 5)."),
				ex("prices = [7,6,4,3,1]", "0", "Prices only fall."),
			},
			constraints: []string{
				"$1 \\le n \\le 10^5$",
				"$0 \\le prices[i] \\le 10^4$",
			},
			samples: []services.SampleTestCases{
				tc("prices = [7,1,5,3,6,4]", "5"),
				tc("prices = [7,6,4,3,1]", "0"),
			},
			hidden: []services.SampleTestCases{
				tc("prices = [1]", "0"),
				tc("prices = [2,4,1]", "2"),
				tc("prices = [3,2,6,5,0,3]", "4"),
			},
		},
		{
			title:      "Merge Two Sorted Lists",
			difficulty: "easy",
			topics:     []string{"Linked List"},
			hints: []string{
				"Use a dummy head node and walk both lists in parallel.",
				"At each step, append the smaller node to the merged list.",
			},
			details: `## Statement

You are given the heads of two sorted linked lists $list1$ and $list2$.

Merge them into a single sorted list and return its head.`,
			examples: []services.Example{
				ex("list1 = [1,2,4], list2 = [1,3,4]", "[1,1,2,3,4,4]", ""),
			},
			constraints: []string{
				"$0 \\le n_1, n_2 \\le 50$",
				"$-100 \\le \\text{Node.val} \\le 100$",
				"Both lists are sorted in non-decreasing order.",
			},
			samples: []services.SampleTestCases{
				tc("list1 = [1,2,4]\nlist2 = [1,3,4]", "[1,1,2,3,4,4]"),
				tc("list1 = []\nlist2 = []", "[]"),
			},
			hidden: []services.SampleTestCases{
				tc("list1 = []\nlist2 = [0]", "[0]"),
				tc("list1 = [-2,3]\nlist2 = [-1,2]", "[-2,-1,2,3]"),
			},
		},
		{
			title:      "Contains Duplicate",
			difficulty: "easy",
			topics:     []string{"Arrays", "Hashing"},
			hints: []string{
				"A set lets you check seen values in $O(1)$.",
			},
			details: `## Statement

Given an integer array $nums$, return $true$ if any value appears at least twice, and $false$ if every element is distinct.`,
			examples: []services.Example{
				ex("nums = [1,2,3,1]", "true", ""),
				ex("nums = [1,2,3,4]", "false", ""),
			},
			constraints: []string{
				"$1 \\le n \\le 10^5$",
				"$-10^9 \\le nums[i] \\le 10^9$",
			},
			samples: []services.SampleTestCases{
				tc("nums = [1,2,3,1]", "true"),
				tc("nums = [1,2,3,4]", "false"),
			},
			hidden: []services.SampleTestCases{
				tc("nums = [1,1,1,3,3,4,3,2,4,2]", "true"),
				tc("nums = [0]", "false"),
			},
		},
		{
			title:      "Move Zeroes",
			difficulty: "easy",
			topics:     []string{"Arrays", "Two Pointers"},
			hints: []string{
				"Use a write pointer for the next non-zero slot.",
				"Walk a read pointer through the array; copy non-zeros forward, then pad with zeros.",
			},
			details: `## Statement

Given an integer array $nums$, move all $0$'s to the end while maintaining the relative order of the non-zero elements.

You must do this **in place** without making a copy of the array.`,
			examples: []services.Example{
				ex("nums = [0,1,0,3,12]", "[1,3,12,0,0]", ""),
			},
			constraints: []string{
				"$1 \\le n \\le 10^4$",
				"$-2^{31} \\le nums[i] \\le 2^{31} - 1$",
			},
			samples: []services.SampleTestCases{
				tc("nums = [0,1,0,3,12]", "[1,3,12,0,0]"),
				tc("nums = [0]", "[0]"),
			},
			hidden: []services.SampleTestCases{
				tc("nums = [1,0,1]", "[1,1,0]"),
				tc("nums = [0,0,0,1]", "[1,0,0,0]"),
				tc("nums = [4,2,4,0,0,3,0,5,1,0]", "[4,2,4,3,5,1,0,0,0,0]"),
			},
		},
		{
			title:      "Valid Anagram",
			difficulty: "easy",
			topics:     []string{"Strings", "Hashing"},
			hints: []string{
				"If lengths differ, they cannot be anagrams.",
				"Count character frequencies and compare.",
			},
			details: `## Statement

Given two strings $s$ and $t$, return $true$ if $t$ is an anagram of $s$, and $false$ otherwise.

An **anagram** uses exactly the same characters with the same frequencies, just rearranged.`,
			examples: []services.Example{
				ex(`s = "anagram", t = "nagaram"`, "true", ""),
				ex(`s = "rat", t = "car"`, "false", ""),
			},
			constraints: []string{
				"$1 \\le |s|, |t| \\le 5 \\cdot 10^4$",
				"$s$ and $t$ contain only lowercase English letters.",
			},
			samples: []services.SampleTestCases{
				tc(`s = "anagram"` + "\n" + `t = "nagaram"`, "true"),
				tc(`s = "rat"` + "\n" + `t = "car"`, "false"),
			},
			hidden: []services.SampleTestCases{
				tc(`s = "a"` + "\n" + `t = "ab"`, "false"),
				tc(`s = "abc"` + "\n" + `t = "cba"`, "true"),
				tc(`s = ""` + "\n" + `t = ""`, "true"),
			},
		},
		{
			title:      "Longest Substring Without Repeating Characters",
			difficulty: "medium",
			topics:     []string{"Strings", "Hashing", "Sliding Window"},
			hints: []string{
				"Use a sliding window with two pointers.",
				"Track the last index seen for each character to advance the left edge.",
			},
			details: `## Statement

Given a string $s$, find the length of the longest substring without repeating characters.`,
			examples: []services.Example{
				ex(`s = "abcabcbb"`, "3", `"abc" has length 3.`),
				ex(`s = "bbbbb"`, "1", `"b" has length 1.`),
			},
			constraints: []string{
				"$0 \\le |s| \\le 5 \\cdot 10^4$",
				"$s$ consists of English letters, digits, symbols, and spaces.",
			},
			samples: []services.SampleTestCases{
				tc(`s = "abcabcbb"`, "3"),
				tc(`s = "pwwkew"`, "3"),
			},
			hidden: []services.SampleTestCases{
				tc(`s = ""`, "0"),
				tc(`s = " "`, "1"),
				tc(`s = "dvdf"`, "3"),
			},
		},
	}
}
