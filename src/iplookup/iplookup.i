/* File : iplookup.i */
%module iplookup
%{
#include "iplookup.h"
#include "ip.h"
#include "struct.h"
#include "iconv_ext.h"

%}

%header %{
#include "config.h"
%}

%include "std_vector.i"
%include "std_string.i"
%include "std_map.i"

%include "iplookup.h"
%include "ip.h"
%include "config.h"
%include "struct.h"
%include "iconv_ext.h"
