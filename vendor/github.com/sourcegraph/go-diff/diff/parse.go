
	"sourcegraph.com/sqs/pbtypes"
// ParseMultiFileDiff parses a multi-file unified diff. It returns an error if parsing failed as a whole, but does its
// best to parse as many files in the case of per-file errors. In the case of non-fatal per-file errors, the error
// return value is null and the Errs field in the returned MultiFileDiff is set.
	return &MultiFileDiffReader{reader: bufio.NewReader(r)}
	reader *bufio.Reader
				return nil, io.EOF
			return fd, nil
			return nil, err
	line, err := readLine(r.reader)
	if err != nil {
		return fd, err
					return fd, nil
			return nil, err
	return fd, nil
	return &FileDiffReader{reader: bufio.NewReader(r)}
	reader *bufio.Reader
		ts := pbtypes.NewTimestamp(*origTime)
		fd.OrigTime = &ts
		ts := pbtypes.NewTimestamp(*newTime)
		fd.NewTime = &ts
// timestamps).
		line, err = readLine(r.reader)
		ts, err := time.Parse(diffTimeParseLayout, parts[1])
			return "", nil, err
	return fmt.Sprintf("overflowed into next file: %s", e)
			line, err = readLine(r.reader)
	switch {
	case (len(fd.Extended) == 3 || len(fd.Extended) == 4 && strings.HasPrefix(fd.Extended[3], "Binary files ")) &&
		strings.HasPrefix(fd.Extended[1], "new file mode ") && strings.HasPrefix(fd.Extended[0], "diff --git "):
		names := strings.SplitN(fd.Extended[0][len("diff --git "):], " ", 2)
		fd.NewName = names[1]
		return true
	case (len(fd.Extended) == 3 || len(fd.Extended) == 4 && strings.HasPrefix(fd.Extended[3], "Binary files ")) &&
		strings.HasPrefix(fd.Extended[1], "deleted file mode ") && strings.HasPrefix(fd.Extended[0], "diff --git "):
		names := strings.SplitN(fd.Extended[0][len("diff --git "):], " ", 2)
		fd.OrigName = names[0]
		return true
	case len(fd.Extended) == 4 && strings.HasPrefix(fd.Extended[2], "rename from ") && strings.HasPrefix(fd.Extended[3], "rename to ") && strings.HasPrefix(fd.Extended[0], "diff --git "):
		names := strings.SplitN(fd.Extended[0][len("diff --git "):], " ", 2)
		fd.OrigName = names[0]
		fd.NewName = names[1]
		return true
	case len(fd.Extended) == 3 && strings.HasPrefix(fd.Extended[2], "Binary files ") && strings.HasPrefix(fd.Extended[0], "diff --git "):
		names := strings.SplitN(fd.Extended[0][len("diff --git "):], " ", 2)
		fd.OrigName = names[0]
		fd.NewName = names[1]
		return true
	default:
		return false
	return &HunksReader{reader: bufio.NewReader(r)}
	reader *bufio.Reader
			line, err = readLine(r.reader)
				// Saw start of new hunk, so this hunk is
				// complete. But we've already read in the next hunk's