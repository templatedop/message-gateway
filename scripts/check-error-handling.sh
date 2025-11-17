#!/bin/bash

# Error Handling Bypass Detection Script
# This script detects when developers bypass the api-errors module
# by directly calling ctx.JSON() with error responses

set -e

echo "🔍 Checking for error handling bypasses..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Initialize counters
VIOLATIONS=0

# Pattern 1: Direct gin.H{"error": ...} usage
echo ""
echo "Checking for Pattern 1: gin.H{\"error\": ...}"
if git grep -n 'gin\.H{.*"error"' -- '*.go' ':!cover.html' ':!*_test.go' ':!vendor/' ':!api-errors/' 2>/dev/null; then
    echo -e "${RED}❌ Found gin.H{\"error\": ...} pattern${NC}"
    ((VIOLATIONS++))
else
    echo -e "${GREEN}✅ No gin.H{\"error\": ...} violations${NC}"
fi

# Pattern 2: ctx.JSON with http.StatusXXX and literal error messages
echo ""
echo "Checking for Pattern 2: ctx.JSON(http.Status..., gin.H{...})"
if git grep -n '\.JSON(http\.Status.*gin\.H{' -- '*.go' ':!cover.html' ':!*_test.go' ':!vendor/' ':!api-errors/' 2>/dev/null | grep -v '// OK: '; then
    echo -e "${RED}❌ Found ctx.JSON with gin.H pattern${NC}"
    ((VIOLATIONS++))
else
    echo -e "${GREEN}✅ No ctx.JSON(status, gin.H{}) violations${NC}"
fi

# Pattern 3: AbortWithStatusJSON with error messages
echo ""
echo "Checking for Pattern 3: AbortWithStatusJSON(..., gin.H{...})"
if git grep -n 'AbortWithStatusJSON.*gin\.H{' -- '*.go' ':!cover.html' ':!*_test.go' ':!vendor/' ':!api-errors/' 2>/dev/null | grep -v 'router.txt' | grep -v '// OK: '; then
    echo -e "${RED}❌ Found AbortWithStatusJSON with gin.H pattern${NC}"
    ((VIOLATIONS++))
else
    echo -e "${GREEN}✅ No AbortWithStatusJSON violations${NC}"
fi

# Pattern 4: Check for commented-out proper error handling
echo ""
echo "Checking for Pattern 4: Commented apierrors.* calls"
if git grep -n '// .*apierrors\.' -- 'handler/*.go' 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Found commented apierrors calls - review needed${NC}"
    echo -e "${YELLOW}   (This might indicate intentional bypass)${NC}"
fi

echo ""
echo "═══════════════════════════════════════════"
if [ $VIOLATIONS -gt 0 ]; then
    echo -e "${RED}❌ Found $VIOLATIONS violation(s)${NC}"
    echo ""
    echo "Error handling bypasses detected!"
    echo ""
    echo "✅ DO use api-errors handlers:"
    echo "   - apierrors.HandleError(ctx, err)"
    echo "   - apierrors.HandleWithMessage(ctx, \"message\")"
    echo "   - apierrors.HandleNotFoundError(ctx)"
    echo "   - apierrors.HandleBadRequestError(ctx)"
    echo ""
    echo "❌ DON'T use direct responses:"
    echo "   - ctx.JSON(status, gin.H{\"error\": msg})"
    echo "   - ctx.AbortWithStatusJSON(status, gin.H{...})"
    echo ""
    echo "See api-errors/ERROR_HANDLING_ANALYSIS.md for details"
    echo "═══════════════════════════════════════════"
    exit 1
else
    echo -e "${GREEN}✅ All checks passed!${NC}"
    echo "═══════════════════════════════════════════"
    exit 0
fi
