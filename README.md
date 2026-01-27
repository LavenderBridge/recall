# Recall

Recall is a CLI-based spaced repetition tool designed to help you master LeetCode problems. It uses the SM-2 algorithm to schedule reviews, ensuring you practice problems effectively over time.

## Installation

```bash
# Clone the repository
git clone https://github.com/LavenderBridge/spaced-repetition.git
cd spaced-repetition

# Build the binary
go build -o recall .

# Optional: Move to your path
sudo mv recall /usr/local/bin/
```

## Usage

### Add a Problem
Start tracking a new problem.
```bash
recall add "Two Sum" 3 --url "https://leetcode.com/problems/two-sum" --tags "array,hashmap" --notes "Watch out for edge cases"
```
*   `3` is the initial difficulty (1-5).
*   `--url`: Link to the problem.
*   `--tags`: Comma-separated tags.
*   `--notes`: Initial notes.

### Review Problems
Start an interactive review session for problems due today.
```bash
recall review
```
*   **Open in Browser**: Automatically open the problem URL before reviewing.
    ```bash
    recall review --open
    ```
*   **Specific Problem**: Review a specific problem by name.
    ```bash
    recall review "Two Sum"
    ```

### List Problems
View all tracked problems.
```bash
recall list
```

### Check Due Problems
See what's due for review today without starting a session.
```bash
recall due
```

### Edit a Problem
Update details for an existing problem.
```bash
recall edit [ID] --difficulty 4 --notes "New improved approach"
```

### Delete a Problem
Remove a problem from tracking.
```bash
recall delete [ID]
```

### Statistics
View your progress distribution.
```bash
recall stats
```

## Algorithm
Recall uses the **SuperMemo-2 (SM-2)** algorithm.
1.  **Quality Rating (0-5)**: You rate your recall quality.
2.  **Interval Calculation**:
    *   `EF' = EF + (0.1 - (5-q) * (0.08 + (5-q)*0.02))`
    *   `Interval = PreviousInterval * EF'`
3.  **Result**: Problems you know well are pushed further into the future; problems you struggle with appear sooner.
