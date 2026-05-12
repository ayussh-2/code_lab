package main

import (
	"errors"
	"flag"
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
	reset := flag.Bool("reset", false, "truncate problem-related tables before seeding")
	flag.Parse()

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

	if *reset {
		if err := resetProblemData(db, log); err != nil {
			log.Fatal("reset failed", zap.Error(err))
		}
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

// resetProblemData wipes everything tied to problems/judging but leaves users
// and OTPs intact. Postgres TRUNCATE ... CASCADE handles FK fan-out for us.
func resetProblemData(db *gorm.DB, log *zap.Logger) error {
	log.Warn("reset flag set: truncating submission_test_results, submissions, test_cases, problems, topics")
	return db.Exec(
		"TRUNCATE submission_test_results, submissions, test_cases, problems, topics RESTART IDENTITY CASCADE",
	).Error
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

// ioFormat renders a consistent "Input / Output Format" markdown block that
// gets appended to every problem's details. Keeps the contract identical
// across the catalogue.
func ioFormat(input, output string) string {
	return "\n\n## Input / Output Format\n\n**Input**\n" + input + "\n\n**Output**\n" + output + "\n"
}

func seedData() []seedProblem {
	return []seedProblem{
		{
			title:      "Echo Input",
			difficulty: "easy",
			topics:     []string{"Implementation"},
			hints: []string{
				"Read from stdin and print the same value.",
			},
			details: `## Statement

Given a line of input, print it unchanged.` + ioFormat(
				"- A single line of text.",
				"- The exact same line.",
			),
			examples: []services.Example{
				ex("hello", "hello", "The output matches the input."),
			},
			constraints: []string{
				"Input length is at most 100 characters.",
			},
			samples: []services.SampleTestCases{
				tc("hello\n", "hello\n"),
			},
			hidden: []services.SampleTestCases{
				tc("world\n", "world\n"),
				tc("42\n", "42\n"),
			},
		},
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

You may assume each input has exactly one solution, and you may not use the same element twice.` + ioFormat(
				"- Line 1: two integers `n` and `target` separated by a space.\n- Line 2: `n` space-separated integers `nums[0] nums[1] ... nums[n-1]`.",
				"- A single line with two space-separated 0-based indices `i j` such that `nums[i] + nums[j] == target`.",
			),
			examples: []services.Example{
				ex("nums = [2,7,11,15], target = 9", "0 1", "nums[0] + nums[1] == 9"),
				ex("nums = [3,2,4], target = 6", "1 2", "nums[1] + nums[2] == 6"),
			},
			constraints: []string{
				"$2 \\le n \\le 10^4$",
				"$-10^9 \\le nums[i] \\le 10^9$",
				"Only one valid answer exists.",
			},
			samples: []services.SampleTestCases{
				tc("4 9\n2 7 11 15", "0 1"),
				tc("3 6\n3 2 4", "1 2"),
			},
			hidden: []services.SampleTestCases{
				tc("2 6\n3 3", "0 1"),
				tc("5 -8\n-1 -2 -3 -4 -5", "2 4"),
				tc("4 0\n0 4 3 0", "0 3"),
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

A string is valid if open brackets are closed by the same type of bracket and in the correct order.` + ioFormat(
				"- A single line containing the string `s`. May be empty.",
				"- `true` or `false` (lowercase).",
			),
			examples: []services.Example{
				ex(`s = "()"`, "true", "Single matched pair."),
				ex(`s = "(]"`, "false", "Mismatched bracket types."),
			},
			constraints: []string{
				"$1 \\le |s| \\le 10^4$",
				"$s$ consists only of bracket characters.",
			},
			samples: []services.SampleTestCases{
				tc("()[]{}", "true"),
				tc("([)]", "false"),
			},
			hidden: []services.SampleTestCases{
				tc("{[]}", "true"),
				tc("(", "false"),
				tc("", "true"),
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

Given the $head$ of a singly linked list, reverse the list and return the new head.` + ioFormat(
				"- Line 1: integer `n`, the number of nodes.\n- Line 2: `n` space-separated integer values from head to tail. May be empty when `n = 0`.",
				"- A single line with the reversed values space-separated. Empty when the list is empty.",
			),
			examples: []services.Example{
				ex("head = [1,2,3,4,5]", "5 4 3 2 1", "Reverse the order of nodes."),
			},
			constraints: []string{
				"$0 \\le n \\le 5000$",
				"$-5000 \\le \\text{Node.val} \\le 5000$",
			},
			samples: []services.SampleTestCases{
				tc("5\n1 2 3 4 5", "5 4 3 2 1"),
				tc("2\n1 2", "2 1"),
			},
			hidden: []services.SampleTestCases{
				tc("0\n", ""),
				tc("1\n7", "7"),
				tc("4\n1 1 2 2", "2 2 1 1"),
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

This is also known as **Kadane's algorithm**.` + ioFormat(
				"- Line 1: integer `n`, the array length.\n- Line 2: `n` space-separated integers.",
				"- A single integer: the maximum contiguous subarray sum.",
			),
			examples: []services.Example{
				ex("nums = [-2,1,-3,4,-1,2,1,-5,4]", "6", "Subarray [4,-1,2,1] has sum 6."),
			},
			constraints: []string{
				"$1 \\le n \\le 10^5$",
				"$-10^4 \\le nums[i] \\le 10^4$",
			},
			samples: []services.SampleTestCases{
				tc("9\n-2 1 -3 4 -1 2 1 -5 4", "6"),
				tc("1\n1", "1"),
			},
			hidden: []services.SampleTestCases{
				tc("5\n5 4 -1 7 8", "23"),
				tc("4\n-1 -2 -3 -4", "-1"),
				tc("3\n0 0 0", "0"),
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

Return the number of distinct ways to reach the top.` + ioFormat(
				"- A single integer `n`.",
				"- A single integer: the number of distinct ways to reach the top.",
			),
			examples: []services.Example{
				ex("n = 2", "2", "1+1 or 2."),
				ex("n = 3", "3", "1+1+1, 1+2, or 2+1."),
			},
			constraints: []string{
				"$1 \\le n \\le 45$",
			},
			samples: []services.SampleTestCases{
				tc("2", "2"),
				tc("3", "3"),
			},
			hidden: []services.SampleTestCases{
				tc("1", "1"),
				tc("5", "8"),
				tc("10", "89"),
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

Your algorithm must run in $O(\log n)$ time.` + ioFormat(
				"- Line 1: two integers `n` and `target`.\n- Line 2: `n` space-separated integers in ascending order.",
				"- A single integer: the 0-based index of `target`, or `-1` if not present.",
			),
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
				tc("6 9\n-1 0 3 5 9 12", "4"),
				tc("6 2\n-1 0 3 5 9 12", "-1"),
			},
			hidden: []services.SampleTestCases{
				tc("1 5\n5", "0"),
				tc("1 -3\n5", "-1"),
				tc("10 1\n1 2 3 4 5 6 7 8 9 10", "0"),
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

Choose a single day to buy and a later day to sell so that profit is maximized. Return the maximum profit, or $0$ if no profit is achievable.` + ioFormat(
				"- Line 1: integer `n`, the number of days.\n- Line 2: `n` space-separated integers, the daily prices.",
				"- A single integer: the maximum profit (`0` if no profit is possible).",
			),
			examples: []services.Example{
				ex("prices = [7,1,5,3,6,4]", "5", "Buy at 1 (day 2), sell at 6 (day 5)."),
				ex("prices = [7,6,4,3,1]", "0", "Prices only fall."),
			},
			constraints: []string{
				"$1 \\le n \\le 10^5$",
				"$0 \\le prices[i] \\le 10^4$",
			},
			samples: []services.SampleTestCases{
				tc("6\n7 1 5 3 6 4", "5"),
				tc("5\n7 6 4 3 1", "0"),
			},
			hidden: []services.SampleTestCases{
				tc("1\n1", "0"),
				tc("3\n2 4 1", "2"),
				tc("6\n3 2 6 5 0 3", "4"),
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

Merge them into a single sorted list and return its head.` + ioFormat(
				"- Line 1: two integers `n1 n2`, the lengths of the two lists.\n- Line 2: `n1` space-separated integers (the first list). May be empty when `n1 = 0`.\n- Line 3: `n2` space-separated integers (the second list). May be empty when `n2 = 0`.",
				"- A single line with the merged values space-separated. Empty when both lists are empty.",
			),
			examples: []services.Example{
				ex("list1 = [1,2,4], list2 = [1,3,4]", "1 1 2 3 4 4", ""),
			},
			constraints: []string{
				"$0 \\le n_1, n_2 \\le 50$",
				"$-100 \\le \\text{Node.val} \\le 100$",
				"Both lists are sorted in non-decreasing order.",
			},
			samples: []services.SampleTestCases{
				tc("3 3\n1 2 4\n1 3 4", "1 1 2 3 4 4"),
				tc("0 0\n\n", ""),
			},
			hidden: []services.SampleTestCases{
				tc("0 1\n\n0", "0"),
				tc("2 2\n-2 3\n-1 2", "-2 -1 2 3"),
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

Given an integer array $nums$, return $true$ if any value appears at least twice, and $false$ if every element is distinct.` + ioFormat(
				"- Line 1: integer `n`, the array length.\n- Line 2: `n` space-separated integers.",
				"- `true` if any value appears more than once, otherwise `false`.",
			),
			examples: []services.Example{
				ex("nums = [1,2,3,1]", "true", ""),
				ex("nums = [1,2,3,4]", "false", ""),
			},
			constraints: []string{
				"$1 \\le n \\le 10^5$",
				"$-10^9 \\le nums[i] \\le 10^9$",
			},
			samples: []services.SampleTestCases{
				tc("4\n1 2 3 1", "true"),
				tc("4\n1 2 3 4", "false"),
			},
			hidden: []services.SampleTestCases{
				tc("10\n1 1 1 3 3 4 3 2 4 2", "true"),
				tc("1\n0", "false"),
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

You must do this **in place** without making a copy of the array.` + ioFormat(
				"- Line 1: integer `n`.\n- Line 2: `n` space-separated integers.",
				"- A single line with the resulting array, space-separated.",
			),
			examples: []services.Example{
				ex("nums = [0,1,0,3,12]", "1 3 12 0 0", ""),
			},
			constraints: []string{
				"$1 \\le n \\le 10^4$",
				"$-2^{31} \\le nums[i] \\le 2^{31} - 1$",
			},
			samples: []services.SampleTestCases{
				tc("5\n0 1 0 3 12", "1 3 12 0 0"),
				tc("1\n0", "0"),
			},
			hidden: []services.SampleTestCases{
				tc("3\n1 0 1", "1 1 0"),
				tc("4\n0 0 0 1", "1 0 0 0"),
				tc("10\n4 2 4 0 0 3 0 5 1 0", "4 2 4 3 5 1 0 0 0 0"),
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

An **anagram** uses exactly the same characters with the same frequencies, just rearranged.` + ioFormat(
				"- Line 1: the string `s`. May be empty.\n- Line 2: the string `t`. May be empty.",
				"- `true` if `t` is an anagram of `s`, otherwise `false`.",
			),
			examples: []services.Example{
				ex(`s = "anagram", t = "nagaram"`, "true", ""),
				ex(`s = "rat", t = "car"`, "false", ""),
			},
			constraints: []string{
				"$1 \\le |s|, |t| \\le 5 \\cdot 10^4$",
				"$s$ and $t$ contain only lowercase English letters.",
			},
			samples: []services.SampleTestCases{
				tc("anagram\nnagaram", "true"),
				tc("rat\ncar", "false"),
			},
			hidden: []services.SampleTestCases{
				tc("a\nab", "false"),
				tc("abc\ncba", "true"),
				tc("\n", "true"),
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

Given a string $s$, find the length of the longest substring without repeating characters.` + ioFormat(
				"- A single line containing the string `s`. May be empty.",
				"- A single integer: the length of the longest substring without repeating characters.",
			),
			examples: []services.Example{
				ex(`s = "abcabcbb"`, "3", `"abc" has length 3.`),
				ex(`s = "bbbbb"`, "1", `"b" has length 1.`),
			},
			constraints: []string{
				"$0 \\le |s| \\le 5 \\cdot 10^4$",
				"$s$ consists of English letters, digits, symbols, and spaces.",
			},
			samples: []services.SampleTestCases{
				tc("abcabcbb", "3"),
				tc("pwwkew", "3"),
			},
			hidden: []services.SampleTestCases{
				tc("", "0"),
				tc(" ", "1"),
				tc("dvdf", "3"),
			},
		},
	}
}
