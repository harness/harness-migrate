# Update User Emails

The `update-users` command allows you to update user emails in exported SCM data before importing it into Harness CODE. This is useful when your source SCM system uses different email addresses than those registered in Harness.

## Usage

```bash
./harness-migrator update-users <user-mapping-file.json> --zipFilePath <exported-zip-file>
```

### Parameters

- `<user-mapping-file.json>`: Path to the JSON file containing user email mappings (required)
- `--zipFilePath`: Path to the exported SCM data zip file (optional, defaults to `harness/harness.zip`)
- `--debug`: Enable debug logging (optional)

## User Mapping File

The user mapping file should be a JSON file that maps original email addresses to new email addresses. For example:

```json
{
  "old-email1@example.com": "new-email1@harness.io",
  "old-email2@example.com": "new-email2@harness.io",
  "old-email3@example.com": "new-email3@harness.io"
}
```

## How It Works

The `update-users` command:

1. Extracts the exported zip file to a temporary directory
2. Finds organization directories and repositories
3. Updates user emails in:
   - Pull request author and commenter information
   - Branch rule bypass user emails
4. Creates an updated zip file that replaces the original

**Note**: The command will overwrite the original zip file. If you want to keep the original, make a backup copy before running the command.

## Example

```bash
# Update user emails in the exported data
./harness-migrator update-users user-mapping.json --zipFilePath exported-data.zip
```

## Next Steps

After updating user emails, you can proceed with importing the data into Harness CODE using the `git-import` command as described in the [git-import documentation](../gitimporter/README.md).
