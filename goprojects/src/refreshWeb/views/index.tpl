<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8" />
		<title>临时刷新授权</title>

		<meta name="description" content="Common form elements and layouts" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />

		<!-- basic styles -->

		<link href="/static/css/bootstrap.min.css" rel="stylesheet" />
		<link rel="stylesheet" href="/static/css/font-awesome.min.css" />

		<!--[if IE 7]>
		  <link rel="stylesheet" href="/static/css/font-awesome-ie7.min.css" />
		<![endif]-->

		<!-- page specific plugin styles -->

		<link rel="stylesheet" href="/static/css/jquery-ui-1.10.3.custom.min.css" />
		<link rel="stylesheet" href="/static/css/chosen.css" />
		<link rel="stylesheet" href="/static/css/datepicker.css" />
		<link rel="stylesheet" href="/static/css/bootstrap-timepicker.css" />
		<link rel="stylesheet" href="/static/css/daterangepicker.css" />
		<link rel="stylesheet" href="/static/css/colorpicker.css" />

		<!-- fonts -->

		<link rel="stylesheet" href="http://fonts.googleapis.com/css?family=Open+Sans:400,300" />

		<!-- ace styles -->

		<link rel="stylesheet" href="/static/css/ace.min.css" />
		<link rel="stylesheet" href="/static/css/ace-rtl.min.css" />
		<link rel="stylesheet" href="/static/css/ace-skins.min.css" />

		<!--[if lte IE 8]>
		  <link rel="stylesheet" href="/static/css/ace-ie.min.css" />
		<![endif]-->

		<!-- inline styles related to this page -->

		<!-- ace settings handler -->

		<script src="/static/js/ace-extra.min.js"></script>

		<!-- HTML5 shim and Respond.js IE8 support of HTML5 elements and media queries -->

		<!--[if lt IE 9]>
		<script src="/static/js/html5shiv.js"></script>
		<script src="/static/js/respond.min.js"></script>
		<![endif]-->
		<style>
		#header {
		    background-color:black;
		    color:white;
		    text-align:center;
		    padding:5px;
		}
		</style>
	</head>
	<body>
	
		<div id="header">
		<h1>临时刷新授权</h1>
		</div>
		<h2> </h2>
		<form id='login-form' class="form-horizontal" method="POST">
	
			<div class="form-group">
				<label class="col-sm-3 control-label no-padding-right" for="form-field-1"> 客户编号 </label>

				<div class="col-sm-9">
					<input type="text" name="custid" placeholder="客户编号" class="col-xs-10 col-sm-5" value="{{.Custid}}" />
				</div>
			</div>

			<div class="space-4"></div>
			
			<div class="form-group">
				<label class="col-sm-3 control-label no-padding-right" for="form-field-1"> 客户证号 </label>

				<div class="col-sm-9">
					<input type="text" name="markno" placeholder="客户证号" class="col-xs-10 col-sm-5" value="{{.Markno}}" />
				</div>
			</div>

			<div class="space-4"></div>
			
			<div class="form-group">
				<label class="col-sm-3 control-label no-padding-right" for="form-field-1"> 设备编号 </label>

				<div class="col-sm-9">
					<input type="text" name="keyno" placeholder="智能卡号/CM地址" class="col-xs-10 col-sm-5" value="{{.Keyno}}"/>
				</div>
			</div>

			<div class="space-4"></div>
			

			<div class="clearfix form-actions">
				<div class="col-md-offset-4 col-md-9">
					<button class="btn btn-info" type="submit">
						&nbsp;提 交&nbsp;
					</button>
					&nbsp; &nbsp; &nbsp; &nbsp;
					<button class="btn" type="reset">
						&nbsp;重 置&nbsp;
					</button>
					
				</div>
			</div>

			<div class="hr hr-24"></div>

		</form>
		
		{{if .Message}}
			<script language="javascript">  
				alert({{.Message}});
			</script>
		{{end}}
	</body>
	
</html>