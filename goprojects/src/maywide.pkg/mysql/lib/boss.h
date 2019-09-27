#include <stdio.h>
#include <stdlib.h>
#include <memory.h>
#include <math.h>
#include <time.h>
#include <string.h>
#include <strings.h>
#include <stdarg.h>
#include <unistd.h>
#include <signal.h>
#include <errno.h>
#include <stddef.h>

/* large file supported, greater than 2g */
#define _FILE_OFFSET_BITS 64
#define __USE_FILE_OFFSET64
#define __USE_LARGEFILE64
#define _LARGEFILE64_SOURCE

#define  SQL_DEBUG    0      /* trace only return (0) */
#define  SQL_CHECK    1      /* return sqlca.sqlcode */
#define  SQL_SELECT   2      /* -1405 ignored, null is valid */
#define  SQL_UPDATE   3      /* 100 or 1403 ignored */
#define  SQL_DELETE   4      /* 100 or 1403 ignored */
#define  SQL_INSERT   5      /* return sqlca.sqlcode */

#ifndef TRUE
  #define  TRUE       1
#endif

#ifndef FALSE
  #define  FALSE      0
#endif

#define CheckTrace(...) \
  CheckError("trace.file.function.rows: %s->%s->%d\n",\
   __FILE__, __FUNCTION__, __LINE__); \
  CheckError(__VA_ARGS__)\

#define SuccessReturn(out) {\
  int node = tab2xml_add(out, -1, "response", "");\
  tab2xml_add(out, node, "status", "0");\
  tab2xml_add(out, node, "code", "0");\
  tab2xml_add(out, node, "message", "Ö´ÐÐ³É¹¦");}\
  return(0)\

#define ErrorReturn(out, status, code, msg) {\
  int node = tab2xml_add(out, -1, "response", "");\
  tab2xml_add(out, node, "status", status);\
  tab2xml_add(out, node, "code", code);\
  tab2xml_add(out, node, "message", msg);}\
  return(strcmp(status,"0")?(-1):0)\

/* -lcom */
void array2argv(char ***argv, char *first, int count, int length);
char **dw_malloc(const int row,  const int col);
char *combine(char *str, char *del, char *arr, int slen, int num);
char *cut(char *src, char *del, char *str, int pos);
char *digit(char *digit, char *str);
char *lower(char *str);
char *lpad(char *dest, char *src, int len, char *str);
char *ltrim(char *str);
char *ltrim_e(char *str, char *mode);
char *nextval(char *str);
char *nextvalb(char *str);
char *ntrim(char *str);
char *ntrim_e(char *str, char *mode);
char *package(char *str, char *arr, int slen, int num);
char *replace(char *src, char *mat, char *rep);
char *rpad(char *dest, char *src, int len, char *str);
char *rtrim(char *str);
char *rtrim_e(char *str, char *mode);
char *sequence(char *seq, char *flag, char type);
char *strmove(char *str, int len);
char *push_word(char *src, char *dst, char *value);
char *popd_word(char *value, char *src, char *sdst, char *edst);
char *sysdate(char *str, char *mode, int diff);
char *time2UTC(char *vtime);
char *time_add(char *vtime, int val, int unit);
char *month_add(char *value, char *YYYYMMDD, int count);
char *to_binary(char *bitstr, int num, int bits);
char *trim(char *str);
char *trim_e(char *str, char *mode);
char *upper(char *str);
char *word(char *word, char *str);
char *ztrim(char *str);
char *ztrim_e(char *str, char *mode);
char getbase64value(char ch);
int base64decode(unsigned char *str, const unsigned char *code);
int base64encode(unsigned char *str, const unsigned char *code);
int chrcnt(const char *string, int letter);
int cutter(char *src, char *str, char *dest, int dlen);
int day_between(char *time1, char *time2);
int field_add(char *buf, char *keyname, char *keyvalue);
int field_del(char *buf, char *keyname);
int field_parse(char *buf, char *arr, int len, int isname);
int field_value(char *keyvalue, char *keyname, char *buf);
int isclosed(int fd, char *buf);
int isGB2312(const char *str);
int isGBK(const char *str);
int isreadable(int fd, int sec, long usec);
int iswriteable(int fd, int sec, long usec);
int readable_timeout(int fd, int sec);
int split(char *src, char *del, char *arr, int slen);
int splitter(char *src, char *del, char *arr, int slen);
int time_between(char *time1, char *time2, int unit);
int to_decimal(char *binary);
int unpackage(char *str, char *arr, int slen);
unsigned char *wstrcpy(unsigned char *dst, const unsigned char *src);
unsigned int wstrlen(const unsigned char *src);
void dw_free(char **arr);
char *toraw(char *value, char *dest);
char *tochar(char *value, char *dest);
int filerows(char *pFile);
int file_time(char *filename);

/* -lparam */
extern char __spid__[16];
extern int  __ivalue_exit__;
char *getparam(char *value, char *name);
char *ivalue(char *value, char *inifile, char *section, char *key);
char *pvalue(char *value, char *name);
int eload(char *filename, char *envname);
int iload(char *inifile, char *section);
int initparam(char *inifile);
int iputenv(char *inifile, char *section);
int opentrc(char *filename);
int pload(char *inifile);
int psave(char *name, char *value);
int put_env(char *filename, char *envname);
int saveparam(char *name, char *value);
int topen(char *filename);
void CheckError(const char *fmt, ...);
void CheckErrorRaw(const char *fmt, ...);
void fcheck(FILE *f, char *fname);
int  tchange();
void CheckRaw(const char *fmt, ...);
int CheckBin(unsigned char *ptr, int len, char *message);
char *concat(char *value, ...);

/* -ldbcom of version 2.2.2 */
extern int __debug_mode__;
int DB_CONNECT(char *, int);
int DB_CHECK();
int DB_SELECT(char *, ...);
int DB_SELECT_A(char *, char *, int);
int DB_SELECT_FIRST(char *, char *, int);
int DB_OPEN_CURSOR(char *);
int DB_FETCH_RECORD(int, ...);
int DB_FETCH_RECORD_A(int, char *, int);
int DB_CLOSE_CURSOR(int);
int DB_EXECUTE(char *);
int DB_BIND_VALUE(char *, int);
int DB_PARALLEL(int);
void DB_RELEASE();
void DB_EXIT();
void DB_COMMIT();
void DB_ROLLBACK();
void DB_PREPARE(char *, ...);
void DB_PREPARE_A(char *, char *, int);
int DB_COL_LENGTH(int id, int len[]);
int DB_COL_NAME(int, char *, int);
int DB_COL_TYPE(int, char *);
int DB_CACHE(char *, char *, int, int);

void GO_TRACE(int islog);

int GO_CURSOR(char *v_sql);
int GO_COLUMNS(int id, char **arr, int arrlen);
int GO_NEXT(int curid, char **arr, int arrlen);
int GO_CLOSE(int curid);

int GO_PREPARE(char *v_sql, char **arr, int arrlen);
int GO_SELECT(char *v_sql, char **arr, int arrlen);

int GO_EXECUTE(char *v_sql);
void GO_LOG(char *str);
void GO_SETDB(char *db);
int DB_SQLERRM(char *msg);
