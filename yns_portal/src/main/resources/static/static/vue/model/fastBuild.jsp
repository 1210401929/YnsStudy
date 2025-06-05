<%@ page language="java" contentType="text/html; charset=UTF-8" pageEncoding="UTF-8"%>
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<html>
<head>
    <meta name=renderer  content=webkit>
	<meta charset="UTF-8">
	<title>自由表单</title>
	<!-- bootstrap -->
	<link rel="stylesheet" href="/bmp_pub/static/bootstrap/bootstrap.css" />
	<!--layui-->
	<link rel="stylesheet" href="/bmp_pub/static/layui2.4.2/css/layui.css" />
	<!--elementui-->
	<link rel="stylesheet" href="/bmp_pub/static/vue/elementui/elementui.css" />
	<!--css-->
	<link rel="stylesheet" href="../../CSS/css/public.css"/>
	<link rel="stylesheet" href="../css/freeForm.css"/>
	<link rel="stylesheet" href="../css/extSkin_blue.css"/>
	<link rel="stylesheet" href="css/fastBuild.css"/>
	<!--jquery-->
	<script type="text/javascript" src="/bmp_pub/static/jquery/jquery-1.10.2.min.js"></script>
	<!--loading-->
	<link rel="stylesheet" href="/bmp_pub/static/jquery-plugin/loading/loading.css">
	<script src="/bmp_pub/static/jquery-plugin/loading/loading.js"></script>
	<!--layui-->
	<script type="text/javascript" src="/bmp_pub/static/layui2.4.2/layui.js"></script>
	<!--common-->
	<script type="text/javascript" src="/bmp_pub/static/pubjs/ycdcommon.js"></script>
	<script type="text/javascript" src="/bmp_pub/static/pubjs/common.js"></script>
	<!--vue-->
	<script type="text/javascript" src="/bmp_pub/static/vue/vue.js"></script>
	<!--elementui-->
	<script type="text/javascript" src="/bmp_pub/static/vue/elementui/elementui.js"></script>
	<!--formcommon-->
	<script type="text/javascript" src="../js/form_common.js"></script>
	<!--主js-->
	<script type="text/javascript" src="fastBuild.js"></script>
</head>
<body>
<div id="fastbuild" class="ww hh">
	<div id="topDiv" class="ww center">
		<div id="topDivLeft" class="hh center left">
			<label id="modelLabel" class="hh center text-b right">模板：</label>
			<el-select v-model="modelType" class="center hh" placeholder="请选择">
				<el-option
						v-for="item in modelOptions"
						:key="item.value"
						:label="item.label"
						:value="item.value">
				</el-option>
			</el-select>
		</div>
		<div id="topDivRight" class="hh center right">
			<el-button class="center right" type="primary" @click="addData">新增</el-button>
			<el-button class="center right" type="primary">删除</el-button>
		</div>
	</div>
	<el-divider></el-divider>
	<div id="centerDiv" class="ww">
		<el-table
				:data="tableData"
				:class="['hh']"
				border
				stripe
				ref="multipleTable"
				:row-class-name="rowClassName"
				tooltip-effect="light"
				highlight-current-row
				@cell-click ="cellClick"
				@row-click ="clickGridRow"
				@select ="selectRow"
				@selection-change="selectionChange"
				style="width: 100%">
			<el-table-column
					type="index"
					width="55">
			</el-table-column>
			<el-table-column v-for="(item,index) in tableColumObj.FORM"
							 :prop="item.prop"
							 :label="item.label"
							 :width="item.width"
							 :column-key="item.prop"
							 show-overflow-tooltip
							 :formatter="stateFormat"
							 :key="index">
				<template slot-scope="scope">
					<el-input v-if="item.editor==='el-input'&&scope.column.columnKey===currColField&&(scope.row.seen||scope.$index===currRowIndex)"
							  :class="['hh']"
							  v-model="scope.row[item.prop]"
							  @blur ="loseInputFcous(scope)">
					</el-input>
					<el-select v-if="item.editor==='el-select'&&scope.column.columnKey===currColField&&(scope.row.seen||scope.$index===currRowIndex)"
							   :class="['ww','hh']"
							   v-model="scope.row[item.prop]"
							   @visible-change="commonFunction(item.visible)"
							   @change="selectChange"
							   placeholder="请选择">
						<el-option
								v-if="(item.group)?false:true"
								v-for="option in codeJson[item.code]"
								:key="option.value"
								:label="option.label"
								:value="option.value">
						</el-option>

						<el-option-group
								v-if="(item.group)?true:false"
								v-for="option in codeJson[item.code]"
								:key="option.label"
								:label="option.label">
							<el-option
									v-for="opt in option.options"
									:key="opt.value"
									:label="opt.label"
									:value="opt.value">
							</el-option>
						</el-option-group>
					</el-select>
					<span v-else v-html="stateFormat(scope.row[item.prop],item)"></span>
				</template>
			</el-table-column>
		</el-table>
	</div>
	<el-divider></el-divider>
	<div id="bottomDiv" class="ww">
		<div id="bottomDivInner" class="hh center">
			<el-button class="center" type="primary">确定</el-button>
			<el-button class="center" type="primary">取消</el-button>
		</div>
	</div>
</div>
</body>
</html>
