class Notes::Page
  attr_reader :template, :layout, :local_path, :public_path, :published_at

  def initialize(attributes = {})
    @template = attributes[:template]
    @layout = attributes[:layout]
    @local_path = attributes[:local_path]
    @public_path = attributes[:public_path]
    @attachments = attributes[:attachments]
    @published_at = attributes[:published_at]
  end

  def attachments
    @attachments ||= []
  end
end
