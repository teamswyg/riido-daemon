#define WIN32_LEAN_AND_MEAN
#define NOMINMAX
#include <windows.h>
#include <bcrypt.h>
#include <ncrypt.h>

#include <array>
#include <cstring>
#include <iomanip>
#include <iostream>
#include <sstream>
#include <stdexcept>
#include <string>

namespace {
constexpr wchar_t keyPrefix[] = L"Riido.CliAuth.DisposableProbe.";
constexpr SECURITY_STATUS notRun = static_cast<SECURITY_STATUS>(0xFFFFFFFFu);

struct Measurement {
    SECURITY_STATUS finalize = notRun;
    SECURITY_STATUS privateExport = notRun;
    SECURITY_STATUS originalSign = notRun;
    SECURITY_STATUS reopenedSign = notRun;
    SECURITY_STATUS childSign = notRun;
    bool properties = false;
    bool publicShape = false;
};

class Key {
public:
    NCRYPT_KEY_HANDLE handle = 0;
    std::wstring name;
    bool owned = false;
    Key() = default;
    Key(const Key&) = delete;
    Key& operator=(const Key&) = delete;
    ~Key() { close(); }
    void close() {
        if (handle != 0) {
            NCryptFreeObject(handle);
            handle = 0;
        }
    }
};

void require(bool condition) {
    if (!condition) throw std::runtime_error("probe_failed");
}

std::wstring freshName() {
    std::array<BYTE, 16> random{};
    require(BCryptGenRandom(nullptr, random.data(), static_cast<ULONG>(random.size()),
                           BCRYPT_USE_SYSTEM_PREFERRED_RNG) == 0);
    std::wostringstream out;
    out << keyPrefix << std::hex << std::setfill(L'0');
    for (BYTE value : random) out << std::setw(2) << static_cast<unsigned>(value);
    return out.str();
}

bool validProbeName(const std::wstring& name) {
    const std::wstring prefix(keyPrefix);
    if (name.size() != prefix.size() + 32 || name.compare(0, prefix.size(), prefix) != 0)
        return false;
    return name.find_first_not_of(L"0123456789abcdef", prefix.size()) == std::wstring::npos;
}

void createKey(NCRYPT_PROV_HANDLE provider, Key& key, bool protectedKey, DWORD exportPolicy) {
    key.name = freshName();
    require(NCryptCreatePersistedKey(provider, &key.handle, NCRYPT_ECDSA_P256_ALGORITHM,
                                    key.name.c_str(), 0, 0) == ERROR_SUCCESS);
    key.owned = true;
    DWORD usage = NCRYPT_ALLOW_SIGNING_FLAG;
    require(NCryptSetProperty(key.handle, NCRYPT_EXPORT_POLICY_PROPERTY,
                             reinterpret_cast<PBYTE>(&exportPolicy), sizeof(exportPolicy), NCRYPT_SILENT_FLAG) == ERROR_SUCCESS);
    require(NCryptSetProperty(key.handle, NCRYPT_KEY_USAGE_PROPERTY,
                             reinterpret_cast<PBYTE>(&usage), sizeof(usage), NCRYPT_SILENT_FLAG) == ERROR_SUCCESS);
    if (protectedKey) {
        NCRYPT_UI_POLICY policy{};
        policy.dwVersion = 1;
        policy.dwFlags = NCRYPT_UI_PROTECT_KEY_FLAG | NCRYPT_UI_FORCE_HIGH_PROTECTION_FLAG;
        policy.pszFriendlyName = L"Riido disposable protection probe";
        require(NCryptSetProperty(key.handle, NCRYPT_UI_POLICY_PROPERTY,
                                 reinterpret_cast<PBYTE>(&policy), sizeof(policy), NCRYPT_SILENT_FLAG) == ERROR_SUCCESS);
    }
}

bool propertyEquals(NCRYPT_KEY_HANDLE key, LPCWSTR property, DWORD expected) {
    DWORD value = 0, size = 0;
    return NCryptGetProperty(key, property, reinterpret_cast<PBYTE>(&value), sizeof(value),
                             &size, NCRYPT_SILENT_FLAG) == ERROR_SUCCESS &&
           size == sizeof(value) && value == expected;
}

bool protectedPolicy(NCRYPT_KEY_HANDLE key) {
    std::array<BYTE, 4096> buffer{};
    DWORD size = 0;
    if (NCryptGetProperty(key, NCRYPT_UI_POLICY_PROPERTY, buffer.data(),
                          static_cast<DWORD>(buffer.size()), &size, NCRYPT_SILENT_FLAG) != ERROR_SUCCESS ||
        size < sizeof(NCRYPT_UI_POLICY) || size > buffer.size()) return false;
    DWORD version = 0, flags = 0;
    std::memcpy(&version, buffer.data(), sizeof(version));
    std::memcpy(&flags, buffer.data() + sizeof(version), sizeof(flags));
    return version == 1 && flags == (NCRYPT_UI_PROTECT_KEY_FLAG | NCRYPT_UI_FORCE_HIGH_PROTECTION_FLAG);
}

bool publicShape(NCRYPT_KEY_HANDLE key) {
    std::array<BYTE, sizeof(BCRYPT_ECCKEY_BLOB) + 64> bytes{};
    DWORD size = 0;
    if (NCryptExportKey(key, 0, BCRYPT_ECCPUBLIC_BLOB, nullptr, bytes.data(),
                        static_cast<DWORD>(bytes.size()), &size, NCRYPT_SILENT_FLAG) != ERROR_SUCCESS ||
        size != bytes.size()) return false;
    BCRYPT_ECCKEY_BLOB header{};
    std::memcpy(&header, bytes.data(), sizeof(header));
    return header.dwMagic == BCRYPT_ECDSA_PUBLIC_P256_MAGIC && header.cbKey == 32;
}

SECURITY_STATUS privateExport(NCRYPT_KEY_HANDLE key) {
    std::array<BYTE, 4096> bytes{};
    DWORD size = 0;
    // A size-only query is not evidence that private material cannot be exported.
    auto status = NCryptExportKey(key, 0, BCRYPT_ECCPRIVATE_BLOB, nullptr,
        bytes.data(), static_cast<DWORD>(bytes.size()), &size, NCRYPT_SILENT_FLAG);
    if (status == ERROR_SUCCESS) {
        BCRYPT_ECCKEY_BLOB header{};
        std::memcpy(&header, bytes.data(), sizeof(header));
        if (size != sizeof(header) + 96 || header.dwMagic != BCRYPT_ECDSA_PRIVATE_P256_MAGIC ||
            header.cbKey != 32) status = NTE_BAD_DATA;
    }
    SecureZeroMemory(bytes.data(), bytes.size());
    return status;
}

bool privateExportDenied(SECURITY_STATUS status) {
    // NTE_NOT_SUPPORTED is observed for nonexportable Software KSP keys. This
    // denial is meaningful only after an identical export succeeds for a control.
    return status == NTE_PERM || status == NTE_NOT_SUPPORTED;
}

SECURITY_STATUS sign(NCRYPT_KEY_HANDLE key) {
    std::array<BYTE, 32> hash{};
    hash.fill(0x23);  // Public synthetic prehash; never a caller-supplied payload.
    std::array<BYTE, 64> signature{};
    DWORD size = 0;
    auto status = NCryptSignHash(key, nullptr, hash.data(), static_cast<DWORD>(hash.size()),
        signature.data(), static_cast<DWORD>(signature.size()), &size, NCRYPT_SILENT_FLAG);
    if (status == ERROR_SUCCESS) {
        status = size == signature.size()
            ? NCryptVerifySignature(key, nullptr, hash.data(), static_cast<DWORD>(hash.size()),
                                    signature.data(), size, 0)
            : NTE_BAD_SIGNATURE;
    }
    SecureZeroMemory(signature.data(), signature.size());
    return status;
}

SECURITY_STATUS reopenAndSign(NCRYPT_PROV_HANDLE provider, const std::wstring& name) {
    Key reopened;
    auto status = NCryptOpenKey(provider, &reopened.handle, name.c_str(), 0, NCRYPT_SILENT_FLAG);
    return status == ERROR_SUCCESS ? sign(reopened.handle) : status;
}

SECURITY_STATUS childSign(const std::wstring& name) {
    std::array<wchar_t, 32768> executable{};
    const DWORD length = GetModuleFileNameW(nullptr, executable.data(), static_cast<DWORD>(executable.size()));
    require(length > 0 && length < executable.size() && validProbeName(name));
    std::wstring command = L"\"" + std::wstring(executable.data(), length) + L"\" --child-sign " + name;
    STARTUPINFOW start{};
    start.cb = sizeof(start);
    PROCESS_INFORMATION process{};
    require(CreateProcessW(executable.data(), command.data(), nullptr, nullptr, FALSE,
                           CREATE_NO_WINDOW, nullptr, nullptr, &start, &process) != FALSE);
    const DWORD wait = WaitForSingleObject(process.hProcess, 10000);
    DWORD code = static_cast<DWORD>(notRun);
    const bool completed = wait == WAIT_OBJECT_0 && GetExitCodeProcess(process.hProcess, &code);
    if (!completed) {
        TerminateProcess(process.hProcess, 1);
        WaitForSingleObject(process.hProcess, 5000);
    }
    CloseHandle(process.hThread);
    CloseHandle(process.hProcess);
    require(completed);
    return static_cast<SECURITY_STATUS>(code);
}

bool cleanup(NCRYPT_PROV_HANDLE provider, Key& key) {
    if (!key.owned) return true;
    if (key.handle != 0 && NCryptDeleteKey(key.handle, NCRYPT_SILENT_FLAG) == ERROR_SUCCESS)
        key.handle = 0;
    key.close();
    Key remaining;
    auto status = NCryptOpenKey(provider, &remaining.handle, key.name.c_str(), 0, NCRYPT_SILENT_FLAG);
    if (status == NTE_BAD_KEYSET) return true;
    if (status != ERROR_SUCCESS) return false;
    status = NCryptDeleteKey(remaining.handle, NCRYPT_SILENT_FLAG);
    if (status != ERROR_SUCCESS) return false;
    remaining.handle = 0;
    status = NCryptOpenKey(provider, &remaining.handle, key.name.c_str(), 0, NCRYPT_SILENT_FLAG);
    return status == NTE_BAD_KEYSET;
}

void measure(NCRYPT_PROV_HANDLE provider, Key& key, bool protectedKey, Measurement& result,
             DWORD exportPolicy = 0) {
    createKey(provider, key, protectedKey, exportPolicy);
    result.finalize = NCryptFinalizeKey(key.handle, NCRYPT_SILENT_FLAG);
    if (result.finalize != ERROR_SUCCESS) return;
    result.properties = propertyEquals(key.handle, NCRYPT_EXPORT_POLICY_PROPERTY, exportPolicy) &&
        propertyEquals(key.handle, NCRYPT_KEY_USAGE_PROPERTY, NCRYPT_ALLOW_SIGNING_FLAG) &&
        propertyEquals(key.handle, NCRYPT_LENGTH_PROPERTY, 256) &&
        (!protectedKey || protectedPolicy(key.handle));
    result.publicShape = publicShape(key.handle);
    result.privateExport = privateExport(key.handle);
    result.originalSign = sign(key.handle);
    key.close();
    result.reopenedSign = reopenAndSign(provider, key.name);
    result.childSign = childSign(key.name);
}

const char* candidateState(const Measurement& result) {
    if (result.finalize == NTE_SILENT_CONTEXT) return "interactive_setup_required";
    if (result.finalize != ERROR_SUCCESS) return "inconclusive_provider_error";
    if (!result.properties || !result.publicShape) return "policy_readback_failed";
    if (result.privateExport == ERROR_SUCCESS) return "private_export_allowed";
    if (result.originalSign == ERROR_SUCCESS || result.reopenedSign == ERROR_SUCCESS ||
        result.childSign == ERROR_SUCCESS) return "silent_use_allowed";
    if (!privateExportDenied(result.privateExport)) return "inconclusive_provider_error";
    if (result.originalSign == NTE_SILENT_CONTEXT && result.reopenedSign == NTE_SILENT_CONTEXT &&
        result.childSign == NTE_SILENT_CONTEXT) return "silent_use_denied_only";
    return "inconclusive_provider_error";
}

void printStatus(SECURITY_STATUS status) {
    if (status == notRun) std::cout << "null";
    else std::cout << static_cast<unsigned long>(static_cast<DWORD>(status));
}

void printMeasurement(const Measurement& result) {
    std::cout << "{\"finalize_status\":"; printStatus(result.finalize);
    std::cout << ",\"properties_match\":" << (result.properties ? "true" : "false");
    std::cout << ",\"public_shape_valid\":" << (result.publicShape ? "true" : "false");
    std::cout << ",\"private_export_status\":"; printStatus(result.privateExport);
    std::cout << ",\"original_sign_status\":"; printStatus(result.originalSign);
    std::cout << ",\"reopened_sign_status\":"; printStatus(result.reopenedSign);
    std::cout << ",\"child_sign_status\":"; printStatus(result.childSign);
    std::cout << "}";
}
}  // namespace

int wmain(int argc, wchar_t** argv) {
    const bool child = argc == 3 && std::wstring(argv[1]) == L"--child-sign" && validProbeName(argv[2]);
    if (!child && (argc != 2 || std::wstring(argv[1]) != L"--ci-silent")) return 2;
    NCRYPT_PROV_HANDLE provider = 0;
    if (NCryptOpenStorageProvider(&provider, MS_KEY_STORAGE_PROVIDER, 0) != ERROR_SUCCESS) return 2;
    if (child) {
        const auto status = reopenAndSign(provider, argv[2]);
        NCryptFreeObject(provider);
        return static_cast<int>(status);
    }
    Key exportControlKey, baselineKey, candidateKey;
    Measurement exportControl, baseline, candidate;
    bool complete = false;
    try {
        measure(provider, exportControlKey, false, exportControl,
                NCRYPT_ALLOW_EXPORT_FLAG | NCRYPT_ALLOW_PLAINTEXT_EXPORT_FLAG);
        require(exportControl.finalize == ERROR_SUCCESS && exportControl.properties &&
                exportControl.publicShape && exportControl.privateExport == ERROR_SUCCESS &&
                exportControl.originalSign == ERROR_SUCCESS && exportControl.reopenedSign == ERROR_SUCCESS &&
                exportControl.childSign == ERROR_SUCCESS);
        measure(provider, baselineKey, false, baseline);
        require(baseline.finalize == ERROR_SUCCESS && baseline.properties && baseline.publicShape &&
                privateExportDenied(baseline.privateExport) && baseline.originalSign == ERROR_SUCCESS &&
                baseline.reopenedSign == ERROR_SUCCESS && baseline.childSign == ERROR_SUCCESS);
        measure(provider, candidateKey, true, candidate);
        complete = true;
    } catch (const std::exception&) {
        // Native diagnostics can contain provider or user context. Emit status metadata only.
    }
    const bool exportControlClean = cleanup(provider, exportControlKey);
    const bool baselineClean = cleanup(provider, baselineKey);
    const bool candidateClean = cleanup(provider, candidateKey);
    NCryptFreeObject(provider);
    std::cout << "{\"version\":1,\"production_supported\":false,\"packaging\":\"unpackaged\","
                 "\"measurement_complete\":" << (complete ? "true" : "false")
              << ",\"cleanup_complete\":" << (exportControlClean && baselineClean && candidateClean ? "true" : "false")
              << ",\"candidate_state\":\"" << candidateState(candidate) << "\",\"export_control\":";
    printMeasurement(exportControl);
    std::cout << ",\"baseline\":";
    printMeasurement(baseline);
    std::cout << ",\"candidate\":"; printMeasurement(candidate);
    std::cout << "}\n";
    return complete && exportControlClean && baselineClean && candidateClean ? 0 : 1;
}
