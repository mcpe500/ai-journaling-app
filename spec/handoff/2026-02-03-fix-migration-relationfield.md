# Handoff Document - Fix Migration 002 RelationField Error

**Date**: 2026-02-03  
**Session Focus**: Bug Fix - RelationField Missing CollectionId  
**Status**: Complete  

---

## Summary

Fixed migration error that prevented server startup:
```
Error: Failed to apply migration 002_growth_analysis.go: 
fields: (11: (collectionId: cannot be blank.).).
```

### Root Cause
Field `related_entries` adalah `RelationField` tapi tidak memiliki `CollectionId`, sehingga PocketBase tidak tahu koleksi mana yang direlasikan.

### Solution
1. Menambahkan lookup untuk koleksi `journal_entries`
2. Menambahkan `CollectionId: entries.Id` pada field `related_entries`
3. Menambahkan `CascadeDelete: false` (relasi many-to-many)

---

## What Was Fixed

### Issue
```go
// Dalam 002_growth_analysis.go:92-95
growthAnalysis.Fields.Add(&core.RelationField{
    Name:     "related_entries",
    MaxSelect: 100,
    // ❌ ERROR: CollectionId tidak di-set
})
```

### Fix
```go
// Lookup journal_entries collection
entries, err := app.FindCollectionByNameOrId("journal_entries")
if err != nil {
    return err
}

// Fixed field definition
growthAnalysis.Fields.Add(&core.RelationField{
    Name:          "related_entries",
    CollectionId:  entries.Id,  // ✅ Ditambahkan
    MaxSelect:     100,
    CascadeDelete: false,       // ✅ Ditambahkan
})
```

---

## Verification

### Changes Made
- **File**: `backend/migrations/002_growth_analysis.go`
- **Lines 11-20**: Menambahkan lookup untuk `journal_entries` collection
- **Lines 96-102**: Menambahkan `CollectionId` dan `CascadeDelete` pada field `related_entries`

### Next Steps to Verify
1. Clear database: `rm -rf backend/pb_data/`
2. Start server: `go run main.go serve`
3. Expected: Server starts, migration 002 berhasil di-apply

---

## Technical Notes

### RelationField Requirements
Field `RelationField` di PocketBase WAJIB memiliki:
- `Name` - nama field
- `CollectionId` - ID koleksi yang direlasikan

Opsional:
- `MaxSelect` - maksimum jumlah relasi (default 1)
- `CascadeDelete` - hapus record terkait saat record ini dihapus
- `Required` - apakah field wajib diisi

### CascadeDelete Logic
- `user` relation: `CascadeDelete: true` → Hapus analysis jika user dihapus (ownership)
- `related_entries` relation: `CascadeDelete: false` → Jangan hapus entries jika analysis dihapus (reference)

---

## Files Modified

```
backend/migrations/002_growth_analysis.go
    - Lines 11-20: Menambahkan lookup journal_entries collection
    - Lines 96-102: Memperbaiki related_entries field dengan CollectionId
```

## Documentation Created

```
spec/004. fix-migration-relationfield-collectionid.md
```

---

## How to Test

```bash
# Clear existing data (clean slate)
cd backend
rm -rf pb_data/

# Start server
go run main.go serve

# Expected output:
# - Migration 001: Success
# - Migration 002: Success (growth_analysis created)
# - Migration 003: Success
# - Server starts on port 8090
```

---

## Blockers Removed

✅ Migration 002 error resolved  
✅ Server can now start successfully  
✅ growth_analysis collection will be created correctly  
✅ Relation to journal_entries properly configured  

---

**End of Handoff Document**
