<script type="text/javascript" src="/static/js/pop.js?ver={{ globalJsVer }}"></script>
<script type="text/javascript" src="/static/js/noticePop.js?ver={{ globalJsVer }}"></script>
<style>
	.copyTextBut {
		display:block;
		margin:0 auto
	}
	/* pop */
	#pop{background:#fff;width:260px;border:1px solid #e0e0e0;font-size:12px;position:fixed;right:10px;bottom:10px;}
	#popHead{line-height:32px;background:#f6f0f3;border-bottom:1px solid #e0e0e0;position:relative;font-size:12px;padding:0 0 0 10px;}
	#popHead h2{font-size:14px;color:#666;line-height:32px;height:32px;}
	#popHead #popClose{position:absolute;right:10px;top:1px;}
	#popHead a#popClose:hover{color:#f00;cursor:pointer;}
	#popContent{padding:5px 10px;}
	#popTitle a{line-height:24px;font-size:14px;font-family:'微软雅黑';color:#333;font-weight:bold;text-decoration:none;}
	#popTitle a:hover{color:#f60;}
	#popIntro{text-indent:24px;line-height:160%;margin:5px 0;color:#666;}
	#popMore{text-align:right;border-top:1px dotted #ccc;line-height:24px;margin:8px 0 0 0;}
	#popMore a{color:#f60;}
	#popMore a:hover{color:#f00;}

	#demo{overflow: hidden;position: absolute;top: 25px;left: 50%;transform: translateX(-50%);}
	#demo td{color:#ffd740;}

	#tip{line-height:32px;background:#f6f0f3;border-bottom:1px solid #e0e0e0;font-size:14px;padding:0 0 0 10px;}
	#tip div.tip-title{color:#666;font-weight:bolder;}
	#tip div.tip-cont{font-size:13px;color:#0000ff;text-indent:2em;}
	#tip div.tip-cont span{color:#ff0000;}

</style>
<div class="left side-menu">
	<div class="sidebar-inner slimscrollleft">
		<div id="sidebar-menu">
			<ul>
            {% for menu in menus %}
                {% if menu.u %}
                <li>
					<a href="{{ menu.u }}" class="waves-effect waves-light" data-nav="">
						<i class="md md-view-list"></i><span>{{ menu.n }}</span>
					</a>
				</li>
                {% else %}
                <li class="has_sub">
					<a href="#" class="waves-effect waves-light"><i class="md md-view-list"></i><span>{{ menu.n }}</span><span class="pull-right"><i class="md md-add"></i></span></a>
					<ul class="list-unstyled">
                    {% for submenu in menu.c %}
						<li><a href="{{ submenu.u }}" data-nav="">{{ submenu.n }}</a></li>
					{%
					endfor %}
                    </ul>
				</li>
                {% endif %}
            {% endfor %}

			</ul>
			<div class="clearfix"></div>
			<div id="tip">
				<div class="tip-title" id="tip-title"></div>
				<div class="tip-cont" id="tip-log"></div>
			</div>
		</div>
		<div class="clearfix"></div>
	</div>
	<div id="pop" style="display:none;">
		<audio id="chatAudio" src="/static/js/pop.wav" type="audio/wav"></audio>
		<audio id="chatAudioBalance" src="/static/js/balance.wav" type="audio/wav"></audio>
		<div id="popHead"> <a id="popClose" title="关闭">关闭</a>
			<h2>温馨提示</h2>
		</div>
		<div id="popContent">
			<dl>
				<dt id="popTitle"><a href="http://blog.csdn.net/xmtblog/">这里是标题</a></dt>
				<dd id="popIntro">这里是内容简介</dd>
			</dl>
			<p id="popMore"><a href="http://blog.csdn.net/xmtblog/">查看 »</a></p>
		</div>
	</div>
</div>
<script>
    $(function() {
        if(window.location.host.slice(0, 8) == 'merchant'){
			common.getAjax("/api/index/tips", function(a) {
			    if(a.success == 0){
			        $("#tip-title").html("温馨提示：");
			        $("#tip-log").html("为了您的资金安全，请绑定<span>谷歌验证器</span>，并定期修改登录和支付密码，最近登录密码修改时间：<span>" + a.loginpwd_log + "</span>，最近支付密码修改时间：<span>" + a.paypwd_log + "</span>");
				}
			});
		}
    });
</script>