package expr

/*
#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <ctype.h>
#include <string.h>
#include <strings.h>

#define   ISTR   1
#define   INUM   2
#define   IOPE   3
#define   ILOG   4
#define   ISUC   0

int  __logic_debug__ = 0;
char *__logic_pos__;
char  __logic_ops__[256];

char *digit(char *nums, char *str)
{
  char *pa = str;

  if (!isdigit(*pa))
    strcpy(nums, "");
  else {
    *nums++ = *pa++;
    while (isdigit(*pa) || *pa == '.')
      *nums++ = *pa++;
    if (*(nums - 1) == '.')
      *(nums - 1) = 0x00;
    else
      *nums = 0x00;
  }
  return(nums);
}

static char *digits(char *nums, char *str)
{
  char *pa = str;

  if (!isdigit(*pa))
    strcpy(nums, "");
  else {
    *nums++ = *pa++;
    while (isdigit(*pa) || *pa == '.')
      *nums++ = *pa++;
    if (*(nums - 1) == '.')
      *(nums - 1) = 0x00;
    else
      *nums = 0x00;
  }
  return(nums);
}

static int get_item(char *expr, char *retstr)
{
  char *p = expr, *st = retstr;

  strcpy(retstr, "");
  if (!*p) return(0);
  if (*p == '>' || *p == '<' || *p == '='
      || *p == '&' || *p == '|' || *p == '!') {
    *st++ = *p++;
    if (*p == '=' || *p == '&' || *p == '|') {
      *st++ = *p++;
    }
    *st = '\0';
    __logic_pos__ = p;
    return(ILOG);
  }
  else if (*p == '(' || *p == ')' || *p == '+'
      || *p == '-' || *p == '*' || *p == '/') {
    *st++ = *p++;
    *st = '\0';
    __logic_pos__ = p;
    return(IOPE);
  }
  else if (isdigit((int)*p)) {
    while (isdigit((int)*p) || *p == '.')
      *st++ = *p++;

    //if str exists, return ISTR
    if (isalnum((int)*p) || *p == '_') {
      while (isalnum((int)*p) || *p == '_')
        *st++ = *p++;
      *st = '\0';
      __logic_pos__ = p;
      return(ISTR);
    }

    *st = '\0';
    __logic_pos__ = p;
    return(INUM);
  }
  else if (isalnum((int)*p)) {
    while (isalnum((int)*p) || *p == '_')
      *st++ = *p++;
    *st = '\0';
    __logic_pos__ = p;
    return(ISTR);
  }
  else if (*p) {
    return(-1);
  }
  return(0);
}

static int func_name(char *name)
{
  if (strcmp(name, "sign") == 0)
    return(1);
  else if (strcmp(name, "abs") == 0)
    return(1);
  else if (strcmp(name, "least") == 0)
    return(1);
  else if (strcmp(name, "greatest") == 0)
    return(1);
  else if (strcmp(name, "trunc") == 0)
    return(1);
  else if (strcmp(name, "ceil") == 0)
    return(1);
  else if (strcmp(name, "round") == 0)
    return(1);
  return(0);
}

static int logic_value(char *val1,  char *val2, char op1, char op2)
{
  if (__logic_debug__)
    printf("logic_value = %s %c%c %s\n", val1, op1, op2, val2);
  switch (op1) {
    case '=': return (strcmp(val1,val2) == 0) ? 1 : 0;
    case '>':
      if (op2 == '=') return (strcmp(val1,val2) >= 0) ? 1 : 0;
      else return (strcmp(val1,val2) > 0) ? 1 : 0;
    case '<':
      if (op2 == '=') return (strcmp(val1,val2) <= 0) ? 1 : 0;
      else return (strcmp(val1,val2) < 0) ? 1 : 0;
    case '!':
      return (strcmp(val1,val2) != 0) ? 1 : 0;
    default: {
      fprintf(stderr, "%s,%s, OPT=%c,%c=\n",
        val1, val2, op1, op2);
      return (-1);
    }
  }
  return (0);
}

static int math_value(double *value, double num1, double num2, char op1, char op2)
{
  if (__logic_debug__)
    printf("math_value = %.4f %c%c %.4f\n", num1, op1, op2, num2);
  switch (op1) {
    case '+': *value = (num1 + num2); return(0);
    case '-': *value = (num1 - num2); return(0);
    case '*': *value = (num1 * num2); return(0);
    case '/': *value = (num1 / num2); return(0);
    case '=': *value = ((num1 == num2) ? 1 : 0); return(0);
    case '>':
      if (op2 == '=') *value = (num1 >= num2) ? 1 : 0;
      else *value = (num1 > num2) ? 1 : 0;
      return(0);
    case '<':
      if (op2 == '=') *value = (num1 <= num2) ? 1 : 0;
      else *value = (num1 < num2) ? 1 : 0;
      return(0);
    case '&': *value = (num1 * num2 != 0) ? 1 : 0; return(0);
    case '|': *value = (num1 != 0 || num2 != 0) ? 1 : 0; return(0);
    case '!': *value = (num1 != num2) ? 1 : 0; return(0);
    default: {
      fprintf(stderr, "%.4f,%.4f, OPT=%c,%c=\n",
        num1, num2, op1, op2);
      return(-1);
    }
  }
  return (0);
}

static int func_value(double *fval, char *name)
{
  int    i;
  double val = 1.0;
  char *rs = NULL, *p;
  int  pn = 1;

  // begin
  i = get_item(__logic_pos__, __logic_ops__);
  if (i != IOPE || __logic_ops__[0] != '(') {
    if (__logic_debug__)
      printf("function without character '('.\n");
    return(-1);
  }

  // next
  for (p = __logic_pos__; *p > 0; p++) {
    if (*p == '(') pn++;
    else if (*p == ')') pn--;
    // found end
    if (0 == pn) break;
  }

  // not found
  if (0 != pn) {
    if (__logic_debug__)
      printf("function without character ')'.\n");
    return(-1);
  }
  rs = p;

  char  func_expr[64], n1[16], n2[16], op1, op2;

  sprintf(func_expr, "%.*s", rs-__logic_pos__, __logic_pos__);
  strcpy(__logic_ops__, rs+1);
  __logic_pos__ = rs+1;

  if (__logic_debug__)
    printf("func.__logic_pos__=%s\n", __logic_pos__);

  if (strcmp(name, "sign") == 0 || strcmp(name, "abs") == 0) {
    if (strchr(func_expr, '+') || strchr(func_expr, '-') > func_expr) {
      if (func_expr[0] == '-') {
        digit(n2, func_expr+1);
        strcpy(n1, "-");
        strcat(n1, n2);
        digit(n2, func_expr+strlen(n1)+1);
        op1 = func_expr[strlen(n1)];
        op2 = '=';
      }
      else {
        digit(n1, func_expr);
        digit(n2, func_expr+strlen(n1)+1);
        op1 = func_expr[strlen(n1)];
        op2 = '=';
      }
      // calc value
      if (math_value(&val, atof(n1), atof(n2), op1, op2) < 0) {
        if (__logic_debug__)
          printf("%s, %s, %c, %c.\n", n1, n2, op1, op2);
        return(-1);
      }
    }
    else
      val = atof(func_expr);
  }
  else if (strcmp(name, "least") == 0 || strcmp(name, "greatest") == 0) {
    rs = strchr(func_expr, ',');
    if (rs == NULL) {
      if (__logic_debug__)
        printf("function without character ','.\n");
      return(-1);
    }

    *rs = 0;
    strcpy(n1, func_expr);
    strcpy(n2, rs + 1);
    if (strcmp(name, "least") == 0) {
      val = (atof(n1) > atof(n2)) ? atof(n2) : atof(n1);
    }
    else if (strcmp(name, "greatest") == 0) {
      val = (atof(n1) > atof(n2)) ? atof(n1) : atof(n2);
    }
    if (__logic_debug__)
      printf("name: %s, %s.\n", name, n1, n2);
  }
  else if (strcmp(name, "round") == 0) {
    p = NULL;
    p = strchr(func_expr, ',');
    if (p == NULL) {
      if (__logic_debug__)
        printf("function without character ','.\n");
      return(-1);
    }

    *p = 0;
    strcpy(n1, func_expr);
    strcpy(n2, p + 1);
    expr_value(&val, func_expr);
    // trunc value for round
    if (atoi(n2) == 0) {
      i = (int)(val*10)%10;
      if (i >= 5)
        *fval = (int)val+1;
      else
        *fval = (int)val;
    }
    else if (atoi(n2) == 1) {
      i = (int)(val*100)%10;
      if (i >= 5)
        *fval = ((int)(val*10)+1)/10.0;
      else
        *fval = ((int)(val*10))/10.0;
    }
    else if (atoi(n2) == 2) {
      i = (int)(val*1000)%10;
      if (i >= 5)
        *fval = ((int)(val*100)+1)/100.0;
      else
        *fval = ((int)(val*100))/100.0;
    }
    else if (atoi(n2) == 3) {
      i = (int)(val*10000)%10;
      if (i >= 5)
        *fval = ((int)(val*1000)+1)/1000.0;
      else
        *fval = ((int)(val*1000))/1000.0;
    }
    else if (atoi(n2) == 4) {
      i = (int)(val*100000)%10;
      if (i >= 5)
        *fval = ((int)(val*10000)+1)/10000.0;
      else
        *fval = ((int)(val*10000))/10000.0;
    }
    else {
      i = (int)(val*10)%10;
      if (i >= 5)
        *fval = (int)val+1;
      else
        *fval = (int)val;
    }
    strcpy(__logic_ops__, rs+1);
    __logic_pos__ = rs+1;
    if (__logic_debug__)
      printf("name: %s, (%s, %s).\n", name, n1, n2);
    return(INUM);
  }
  else {
    // other func-value.
    expr_value(&val, func_expr);
    strcpy(__logic_ops__, rs+1);
    __logic_pos__ = rs+1;
  }

  if (strcmp(name, "sign") == 0) {
    if (val > 0)
      *fval = 1.0;
    else if (val < 0)
      *fval = -1.0;
    else
      *fval = 0.0;
    return(INUM);
  }
  else if (strcmp(name, "abs") == 0) {
    if (val >= 0)
      *fval = val;
    else
      *fval = -1.0 * val;
    return(INUM);
  }
  else if (strcmp(name, "trunc") == 0) {
    *fval = (int)val;
    return(INUM);
  }
  else if (strcmp(name, "ceil") == 0) {
    int l = val * 100.0;
    int r = (int)val*100;

    if (l > r)
      *fval = (int)val + 1;
    else
      *fval = (int)val;
    return(INUM);
  }
  *fval = val;
  return(INUM);
}

static int get_field(double *result, char *retval)
{
  int    r1, r2, r;
  char   op1, op2;
  double num1, num2, value;
  char   val1[64], val2[64];

  r1 = get_item(__logic_pos__, __logic_ops__);
  if (r1 == IOPE && __logic_ops__[0] == '(') {
    r1 = get_value(&num1);
    *result = num1;
    return(r1);
  }
  else if (r1 == ISTR) {
    strcpy(val1, __logic_ops__);
    strcpy(retval, val1);
    if (func_name(val1)) {
      r = func_value(&num1, val1);
      *result = num1;
      r1 = r;
    }
    else
      return(r1);
  }
  else if (r1 == INUM) {
    strcpy(val1, __logic_ops__);
    num1 = atof(__logic_ops__);
  }

  if (r1 == IOPE && __logic_ops__[0] == '-') {
    r1 = get_item(__logic_pos__, __logic_ops__);
    if (r1 == INUM) {
      num1 = -atof(__logic_ops__);
      sprintf(val1, "-%s", __logic_ops__);
      if (__logic_debug__)
        printf("minus found. __logic_ops__=%s\n",
          __logic_ops__);
    }
    else if (r1 == ISTR)
      return(-1);
    else if (r1 == IOPE && __logic_ops__[0] == '(') {
      r1 = get_value(&num1);
      num1 = -1.0 * num1;
      *result = num1;
      return(r1);
    }
  }

loop_factor:
  r = get_item(__logic_pos__, __logic_ops__);
  if (r == IOPE && (__logic_ops__[0] == '+'
    || __logic_ops__[0] == '-' || __logic_ops__[0] == ')')) {
    __logic_pos__--;
    strcpy(retval, val1);
    *result = num1;
    return(r1);
  }
  else if (r == ISUC) {
    *result = num1;
    return(r);
  }

  op1 = __logic_ops__[0];
  op2 = __logic_ops__[1];

  r2 = get_item(__logic_pos__, __logic_ops__);

  if (r2 == INUM) {
    num2 = atof(__logic_ops__);
    strcpy(val2, __logic_ops__);
  }
  else if (r2 == ISTR) {
    strcpy(val2, __logic_ops__);
    if (func_name(val2)) {
      r2 = func_value(&num2, val2);
    }
  }
  else if (r2 == IOPE && __logic_ops__[0] == '(')
     r2 = get_value(&num2);
  if (r2 == IOPE && __logic_ops__[0] == '-') {
    r2 = get_item(__logic_pos__, __logic_ops__);
    if (r2 == INUM) {
      num2 = -atof(__logic_ops__);
      sprintf(val2, "-%s", __logic_ops__);
      if (__logic_debug__)
        printf("minus found. __logic_ops__=%s\n",
          __logic_ops__);
    }
    else if (r2 == ISTR)
      return(-1);
    else if (r2 == IOPE && __logic_ops__[0] == '(') {
      r2 = get_value(&num2);
      num2 = -1.0 * num2;
      *result = num2;
      return(r2);
    }
  }

  if (r1 == ISTR || r2 == ISTR) {
    r = logic_value(val1, val2, op1, op2);
    if (-1 == r)
      return(-1);
    num1 = r;
  }
  else {
    r = math_value(&value, num1, num2, op1, op2);
    if (-1 == r)
      return(-1);
    num1 = value;
  }
  goto loop_factor;
}

int get_value(double *result)
{
  double  num1, num2, value;
  char    val1[64], val2[64];
  char    op1, op2;
  int     r1, r, r2;

  // first
  r1 = get_field(&num1, val1);
  if (-1 == r1)
    return(-1);
  else if (0 == r1) {
    *result = num1;
    return(0);
  }

loop_value :
  // operator
  r = get_item(__logic_pos__, __logic_ops__);
  if (-1 == r)
    return(-1);
  else if (0 == r) {
    *result = num1;
    return(0);
  }

  if (r == IOPE && __logic_ops__[0] == ')') {
    *result = num1;
    return(INUM);
  }

  op1 = __logic_ops__[0];
  op2 = __logic_ops__[1];

  r2 = get_field(&num2, val2);
  // call and return
  if (r1 == ISTR || r2 == ISTR) {
    r = logic_value(val1, val2, op1, op2);
    if (-1 == r)
      return(-1);
    num1 = r;
  }
  else {
    r = math_value(&value, num1, num2, op1, op2);
    if (-1 == r)
      return(-1);
    num1 = value;
  }
goto loop_value;

  return(0);
}

static int expr_build(char *nexpr, char *oexpr, char tag, char *invar, int inlen)
{
  char *p = oexpr, *pa = nexpr;
  char id[8];

  __logic_pos__ = nexpr;
  if (0 == tag || NULL == invar || 0 == inlen) {
    while (*p != 0x00) {
      if (*p > 0x20)
        *pa++ = *p;
      p++;
    }
    *pa = '\0';
    return(0);
  }

  while (*p != 0x00) {
    if (*p <= 0x20) {
      p++;
      continue;
    }

    if (*p == tag && *(p+1) >= 0x30 && *(p+1) <= 0x39) {
      p++;
      digits(id, p);
      if (strlen(id) >= 8)
        return(-1);
      strcpy(pa, invar + inlen * atoi(id));
      if (strchr(invar + inlen * atoi(id), 0x20) || strlen(invar + inlen * atoi(id)) == 0) {
        printf("field %d is null or has blank is not allowed.\n", atoi(id));
        return(-1);
      }
      if (__logic_debug__)
        printf("Field of %c, id=%d replaced, value=%s.\n",
          tag, atoi(id), invar + inlen * atoi(id));
      pa += strlen(invar + inlen * atoi(id));
      p  += strlen(id);
    }
    else {
      *pa++ = *p++;
    }
  }

  *pa = '\0';
  return(0);
}

int expr(double *value, char *expr, char tag, char *invar, int inlen)
{
  char newexpr[256];
  int  i;

  i = expr_build(newexpr, expr, tag, invar, inlen);
  if (i < 0)
    return(-1);
  if (__logic_debug__)
    printf("expr=%s\n", newexpr);
  i = get_value(value);
  return(i);
}

int expr_value(double *value, char *expr)
{
  char newexpr[256];
  int  i;

  char *p = expr, *pa = newexpr;
  char id[8];

  __logic_pos__ = newexpr;
  while (*p != 0x00) {
    if (*p > 0x20)
      *pa++ = *p++;
    else
      p++;
  }
  *pa = '\0';

  if (__logic_debug__)
    printf("expr=%s\n", newexpr);
  i = get_value(value);
  return(i);
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

func Logic_value(expr string) (err error, isok bool) {
	c_expr := C.CString(expr)
	defer C.free(unsafe.Pointer(c_expr))
	d_isok := C.double(0) //default failure

	rr := C.expr_value(&d_isok, c_expr)
	if rr != 0 {
		err = errors.New("expr failed")
		isok = false
		return
	}

	if d_isok == 1 {
		isok = true
	} else {
		isok = false
	}

	return
}

func Expr_value(expr string) (err error, val float32) {
	c_expr := C.CString(expr)
	defer C.free(unsafe.Pointer(c_expr))
	d_val := C.double(0) //default failure

	rr := C.expr_value(&d_val, c_expr)
	if rr != 0 {
		err = errors.New("expr failed")
		val = 0.0
		return
	}
	val = float32(d_val)
	return
}
