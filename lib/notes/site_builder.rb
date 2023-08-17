class Notes::SiteBuilder
  include Notes::Logging

  def build
    site.pages.each { render(_1) }
  end

  private

  def site
    @site ||= Notes::Site.new
  end

  def render(page)
    local_path = page.local_path
    logger.info("rendering #{local_path}")
    locals = {configuration: Notes::Configuration, site: site, page: page}
    content = render_page(template: page.template, layout: page.layout, locals: locals)
    write(local_path: local_path, content: content)
    copy_attachments(local_path: local_path, attachments: page.attachments)
  end

  def write(local_path:, content:)
    path = File.join(build_path, local_path)
    FileUtils.mkdir_p(File.dirname(path))
    File.write(path, content)
  end

  def copy_attachments(local_path:, attachments:)
    attachments.each do |attachment|
      logger.info("attaching #{attachment}")
      cached_file = File.join(assets_path, attachment)
      dest_path = File.join(build_path, File.dirname(local_path), File.basename(attachment))
      next if File.exist?(dest_path) && FileUtils.compare_file(cached_file, dest_path)
      FileUtils.cp_r(cached_file, dest_path, remove_destination: true)
    end
  end

  def render_page(template:, layout:, locals:)
    content = render_template(template: template, locals: locals)
    layout ? render_template(template: layout, locals: locals) { content } : content
  end

  def render_template(template:, locals:, &block)
    load_template(template).render(nil, locals.transform_keys(&:to_s), &block)
  end

  def load_template(name)
    @templates ||= {}
    @templates[name] ||= Tilt::ERBTemplate.new(template_path(name))
  end

  def template_path(template_name)
    File.join(templates_path, "#{template_name}.erb")
  end

  def build_path
    Notes::Configuration.build_path
  end

  def templates_path
    Notes::Configuration.templates_path
  end

  def assets_path
    Notes::Configuration.assets_path
  end
end
