#!/usr/bin/env python3
"""Offline, synthetic four-shell completion matrix; no login or product API.

Requires all four shell executables. Bash/zsh exercise their callbacks with
shell variables; fish and PowerShell use their native completion entry points.
This does not replace interactive TAB or Windows OS acceptance.
"""
from pathlib import Path
import argparse
import subprocess,json,os,shlex,hashlib,sys,shutil

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--binary', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--bash', default=shutil.which('bash'))
parser.add_argument('--zsh', default=shutil.which('zsh'))
parser.add_argument('--fish', default=shutil.which('fish'))
parser.add_argument('--powershell', default=shutil.which('pwsh'))
options = parser.parse_args()
for shell in ['bash', 'zsh', 'fish', 'powershell']:
    if not getattr(options, shell):
        parser.error('All four shells are required; missing ' + shell)

out=options.output.resolve();out.mkdir(mode=0o700)
binary=options.binary.resolve();scripts=out/'scripts';scripts.mkdir()
for shell,name in [('bash','riido.bash'),('zsh','_riido'),('fish','riido.fish'),('powershell','riido.ps1')]:
 (scripts/name).write_bytes(subprocess.check_output([str(binary),'completion',shell]))
shim=out/'shim';shim.mkdir();trace=out/'argv.txt'
(shim/'riido').write_text('#!/bin/sh\nprintf "%s\\n" "$@" >> '+shlex.quote(str(trace))+'\ntee -a '+shlex.quote(str(out/'stdin.bin'))+' | '+shlex.quote(str(binary))+' "$@"\n');(shim/'riido').chmod(0o755)
(out/'synthetic-secret-file').write_text('file fallback must not appear')
cases=[
 ('root','riido wo',['work']),('leaf','riido work ge',['get']),
 ('trailing','riido work ',['get','search']),
 ('profile','riido config list --profile st',['staging']),
 ('equals','riido config list --profile=st',['--profile=staging']),
 ('quoted','riido work search --query "한글 여러 단어" --fo',['--format']),
 ('opaque','riido work get --team-id synt',[]),
 ('secret','riido auth login --token completion-secret-fixture',[]),
 ('hidden','riido dae',[]),('unsupported','riido auth login --with',[]),
 ('keys','riido config set ti',['timeout']),
 ('config-value','riido config set format j',['json']),
 ('empty-value','riido config set workspace "" --fo',['--format']),
 ('quoted-enum','riido config list --profile "st',['staging']),
 ('middle','riido config list --profile prodBAD',['production']),
 ('secret-in-value','riido work search --query --token --format j',[]),
 ('variables','riido work search --query $RIIDO_COMPLETION_SECRET --f',[]),
 ('wildcard','riido work search --query * --fo',['--format']),
 ('substitution','riido work search --query $(touch should-not-exist) --fo',[]),
]
# Complete prefix before BAD, preserving the suffix for PowerShell/Readline.
env={'HOME':str(out),'PATH':str(shim)+os.pathsep+os.defpath,'LC_ALL':'en_US.UTF-8','TERM':'xterm','RIIDO_CLI_PROFILE':'must-not-read-config','RIIDO_COMPLETION_SECRET':'private-expanded-fixture','DOTNET_EnableDiagnostics':'0','POWERSHELL_TELEMETRY_OPTOUT':'1'}
rows=[]
def run(shell,script):
 executable=getattr(options, shell)
 args=[executable,'-c',script] if shell not in ['bash','zsh','powershell'] else ([executable,'--noprofile','--norc','-c',script] if shell=='bash' else [executable,'-f','-c',script] if shell=='zsh' else [executable,'-NoLogo','-NoProfile','-NonInteractive','-Command',script])
 guard=['/usr/bin/sandbox-exec','-p','(version 1)(allow default)(deny network-outbound)'] if Path('/usr/bin/sandbox-exec').exists() else []
 return subprocess.run(guard+args,text=True,capture_output=True,env=env,cwd=out,timeout=30)
for shell in ['bash','zsh','fish','powershell']:
 for name,line,expected in cases:
  prefix=line[:-3] if name=='middle' else line
  # shlex is test-fixture construction only; real scripts obtain words from their shell.
  lexline=prefix+'"' if name=='quoted-enum' else prefix
  words=shlex.split(lexline);current='' if prefix.endswith(' ') else words[-1];before=words[1:] if prefix.endswith(' ') else words[1:-1]
  if shell=='bash':
   bashcurrent=current.split('=',1)[1] if name=='equals' else current
   script='source '+shlex.quote(str(scripts/'riido.bash'))+'; COMP_LINE='+shlex.quote(line)+'; COMP_POINT='+str(len(prefix.encode()))+'; COMP_CWORD=0; COMP_WORDS=('+shlex.quote(bashcurrent)+'); _riido_completion; printf "%s\\n" "${COMPREPLY[@]}"'
   want=['staging'] if name=='equals' else expected
  elif shell=='zsh':
   script='source '+shlex.quote(str(scripts/'_riido'))+'; words=(riido '+ ' '.join(shlex.quote(w) for w in before+[current])+'); CURRENT='+str(len(before)+2)+'; PREFIX='+shlex.quote(current)+'; compadd(){ shift; printf "%s\\n" "$@"; }; _riido'
   want=['--format'] if name in ['variables','substitution'] else expected
  elif shell=='fish':
   script='source '+shlex.quote(str(scripts/'riido.fish'))+'; complete -C '+shlex.quote(prefix)
   want=expected
  else:
   quote=lambda x:"'"+x.replace("'","''")+"'"
   script='. '+quote(str(scripts/'riido.ps1'))+'; (TabExpansion2 -inputScript '+quote(line)+' -cursorColumn '+str(len(prefix))+').CompletionMatches | ForEach-Object { $_.CompletionText }'
   # PowerShell's public native API retains local filename fallback.
   want=['./synthetic-secret-file'] if name=='opaque' else expected
  r=run(shell,script);actual=[x.split('\t')[0] for x in r.stdout.splitlines() if x]
  passed=r.returncode==0 and (all(x in actual for x in want) if want else not actual)
  rows.append({'shell':shell,'case':name,'expected':want,'actual':actual,'exit':r.returncode,'passed':passed,'stderr':r.stderr})
  if not passed:print(shell,name,'FAIL',r.returncode,repr(actual),repr(r.stderr[:160]))
raw=trace.read_text().splitlines() if trace.exists() else []
argv_ok=all(raw[i:i+2]==['completion','__complete'] for i in range(0,len(raw),2))
receipt={'rows':rows,'passes':sum(x['passed'] for x in rows),'cases':len(rows),'helper_calls':len(raw)//2,'helper_argv_only_protocol':argv_ok,'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'network':'OS denied outbound' if Path('/usr/bin/sandbox-exec').exists() else 'not OS guarded','product_api_calls':'not requested','signed_account_acceptance':False}
(out/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
print(receipt['passes'],'/',len(rows),'passed; helper argv protocol-only',argv_ok)
raise SystemExit(0 if all(x['passed'] for x in rows) and argv_ok and not (out/'should-not-exist').exists() and b'private-expanded-fixture' not in (out/'stdin.bin').read_bytes() else 1)
