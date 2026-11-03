# Riido completion for PowerShell 7+. Dot-source this file in the current shell.
Register-ArgumentCompleter -Native -CommandName riido -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    try {
        # Reparse only the prefix at the cursor. Parsing never evaluates input.
        $length = [Math]::Min($cursorPosition - $commandAst.Extent.StartOffset, $commandAst.Extent.Text.Length)
        $prefix = $commandAst.Extent.Text.Substring(0, $length)
        $parsed = [System.Management.Automation.Language.Parser]::ParseInput($prefix, [ref]$null, [ref]$null)
        $command = $parsed.Find({ param($node) $node -is [System.Management.Automation.Language.CommandAst] }, $false)
        $tokens = [System.Collections.Generic.List[string]]::new()
        $tokens.Add('words')
        foreach ($element in $command.CommandElements | Select-Object -Skip 1) {
            # Never evaluate variables, expressions, substitutions or script blocks.
            if ($element -isnot [System.Management.Automation.Language.StringConstantExpressionAst]) { return }
            $tokens.Add($element.Value)
        }
        if ($wordToComplete -eq '' -or $command.CommandElements.Count -le 1 -or $command.CommandElements[-1].Extent.EndOffset -lt $prefix.Length) {
            $tokens.Add('')
        }
        $OutputEncoding = [System.Text.UTF8Encoding]::new($false)
        $inputText = ($tokens -join [char]0) + [char]0
        $candidates = @($inputText | & riido completion __complete 2>$null)
        foreach ($candidate in $candidates) {
            if ($candidate) {
                [System.Management.Automation.CompletionResult]::new($candidate, $candidate, 'ParameterValue', $candidate)
            }
        }
        # No match uses PowerShell's normal native filename fallback. Its public
        # native completer API has no handled-empty result; do not throw or insert
        # a fake candidate to suppress it. Riido itself never returns filenames.
    } catch { return }
}
