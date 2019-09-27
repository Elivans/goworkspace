/**
 * @file expr.c
 *
 * @brief 规则运算器:
 * @brief     支持逻辑、算术、混合、优先级、智能字符串识别
 * @brief     支持可扩展函数计算，目前仅支持sign,abs等函数。
 *
 * @author maywide\@revenco.com
 * @defgroup expr -lexpr高级版规则运算
 * @{
 */

#ifdef __cplusplus
extern "C" {
#endif

#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <ctype.h>
#include <string.h>
#include <strings.h>

#define   ISTR   1   /* 字符串 */
#define   INUM   2   /* 数字 */
#define   IOPE   3   /* 算术计算 */
#define   ILOG   4   /* 逻辑判断 */
#define   ISUC   0   /* 结束 */

int  __logic_debug__ = 0;
char *__logic_pos__;
char  __logic_ops__[256];

char *digit(char *digit, char *str)
{
  char *pa = str;

  if (!isdigit(*pa))
    strcpy(digit, "");
  else {
    *digit++ = *pa++;
    while (isdigit(*pa) || *pa == '.')
      *digit++ = *pa++;
    if (*(digit - 1) == '.')
      *(digit - 1) = 0x00;
    else
      *digit = 0x00;
  }
  return(digit);
}

static char *digits(char *digit, char *str)
{
  char *pa = str;

  if (!isdigit(*pa))
    strcpy(digit, "");
  else {
    *digit++ = *pa++;
    while (isdigit(*pa) || *pa == '.')
      *digit++ = *pa++;
    if (*(digit - 1) == '.')
      *(digit - 1) = 0x00;
    else
      *digit = 0x00;
  }
  return(digit);
}

/**
 * @brief 获取指针开始的表达式内容域
 * @param expr   入参：表达式指针
 * @param retstr 出参：表达式对应操作域内容
 * @return
 * @retval  0  成功
 * @retval -1  失败
 */
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
  // 预留其他函数扩展
  return(0);
}

/**
 * @brief 计算逻辑表达式的结果
 * @param val1 字符串1
 * @param val2 字符串2
 * @param op1  比较符1
 * @param op2  比较符2
 * @return
 * @retval  0  FALSE
 * @retval  1  TRUE
 * @retval -1  计算错误
 */
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

/**
 * @brief 计算算术表达式的结果
 * @param value 算术运算结果输出
 * @param num1 字符串1
 * @param num2 字符串2
 * @param op1  比较符1
 * @param op2  比较符2
 * @return
 * @retval  0  计算成功
 * @retval -1  计算错误
 */
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

  // 识别开始
  i = get_item(__logic_pos__, __logic_ops__);
  if (i != IOPE || __logic_ops__[0] != '(') {
    if (__logic_debug__)
      printf("function without character '('.\n");
    return(-1);
  }

  // 获取下一个右括号
  for (p = __logic_pos__; *p > 0; p++) {
    if (*p == '(') pn++;
    else if (*p == ')') pn--;
    // 如果刚好pn = -1，就是找到右括号
    if (0 == pn) break;
  }
  
  // 找不到就报错
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
      // 计算结果
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
    // 进行round坚持，暂时支持4位够了
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
    // 其他函数，直接计算函数里面表达式的值.
    expr_value(&val, func_expr);
    strcpy(__logic_ops__, rs+1);
    __logic_pos__ = rs+1;
  }
  /*
  // 识别函数值
  i = get_item(__logic_pos__, __logic_ops__);
  if (i == IOPE && __logic_ops__[0] == '-') {
    val = -1.0;
    i = get_item(__logic_pos__, __logic_ops__);
  }
  else if (i != INUM) {
    if (__logic_debug__)
      printf("function.argv isnot value.\n");
    return(-1);
  }

  // 函数值计算
  val *= atof(__logic_ops__);

  // 识别结束符
  i = get_item(__logic_pos__, __logic_ops__);
  if (i != IOPE || __logic_ops__[0] != ')') {
    if (__logic_debug__)
      printf("function without character ')'.\n");
    return(-1);
  }
  */
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
    *fval = ceil(val);
    return(INUM);
  }
  *fval = val;
  // 预留其他函数扩展
  return(INUM);
}

/**
 * @brief 表达式的域解析
 * @param  result 返回算术域结果
 * @param  retval 返回字符串
 * @return
 * @retval   0  成功
 * @retval  -1  失败
 */
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

/**
 * @brief 逻辑表达式计算
 * @param result  出参：逻辑表达式计算结果:[0-false;1true]
 * @return
 * @retval  0 表达式运算成功
 * @retval -1 表达式运算失败
 */
int get_value(double *result)
{
  double  num1, num2, value;
  char    val1[64], val2[64];
  char    op1, op2;
  int     r1, r, r2;

  // 获取第一个值
  r1 = get_field(&num1, val1);
  if (-1 == r1)
    return(-1);
  else if (0 == r1) {
    *result = num1;
    return(0);
  }

loop_value :
  // 获取操作符
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
  // 进行计算并返回结果
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

/**
 * @brief 初始化表达式，进行数组变量绑定
 * @param nexpr  出参：初始化后的可计算表达式
 * @param oexpr  入参：原始
 * @param nexpr  出参：初始化后的可计算表达式
 * @param nexpr  出参：初始化后的可计算表达式
 */
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

/**
 * @brief 智能规则预算.
 *  支持算术运算，加减乘除，乘除优先等规则
 *  支持逻辑运算，支持或非操作
 *  支持符合运算，逻辑与算术混合运算功能
 * @param value  返回值(逻辑表达式时,1成功,0失败)
 * @param expr   表达式逻辑表达式支持strcmp字符串对比.
 * @param tag    自动识别变量索引
 * @param invar  自动替换变量值
 * @param inlen  变量指针长度
 * @return
 * @retval 0 成功
 * @retval 1 失败
 * @par 示例：
 *      该例子自动根据V数组将值绑定到F开头的变量并做运算
 *      无需变量绑定时：expr(&val, "10+2*3-(5>F01)", 0, NULL, 0);
 * @code
    double val;
    char   V[64][64];

    strcpy(V[1], "1");
    i = expr(&val, "10+2*3-(5>F01)", 'F', V[0], sizeof(V[0]));
    if (i < 0)
      printf("expr failed.\n");
    else
      printf("value = %.0f.\n", val);
 * @endcode
 */
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

/**
 * @brief 无替换规则运算
 *        expr函数的简化版
 * @param value  返回值(逻辑表达式时,1成功,0失败)
 * @param expr   表达式逻辑表达式支持strcmp字符串对比.
 * @return
 * @retval 0 成功
 * @retval 1 失败
 * @par 示例：
 *      该例子自动根据V数组将值绑定到F开头的变量并做运算
 *      无需变量绑定时：expr(&val, "10+2*3-(5>F01)", 0, NULL, 0);
 * @code
    double val;
    char   V[64][64];

    strcpy(V[1], "1");
    i = expr_value(&val, "10+2*3-(5>1)");
    if (i < 0)
      printf("expr failed.\n");
    else
      printf("value = %.0f.\n", val);
 * @endcode
 */
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
#ifdef __cplusplus
}
#endif

/** @} */
/*
int main(int argc, char **argv)
{
  if (argc < 3) {
    printf("Usage: %s <expr> <,param,,,>\n", argv[0]);
    exit(1);
  }

  char V[256][256], nexpr[256];
  int  i, ret;

  memset(V, 0, sizeof(V));
  for (i = 0; i < argc-3; i++)
    strcpy(V[i], argv[i+3]);

  double result;
  i = expr(&result, argv[1], argv[2][0], V[0], sizeof(V[0]));
  if (i < 0) {
    printf("*** get_value failed.***\n", i);
  }
  else {
    printf("%s=%.0f\n", argv[1], result);
  }
  return(0);
}
*/
